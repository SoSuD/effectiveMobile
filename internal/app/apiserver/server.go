// Package apiserver provides the HTTP API server with Swagger documentation.
// @title EffectiveMobile API
// @version 1.0
// @description API server for EffectiveMobile service
// @host localhost:8080
// @BasePath /
package apiserver

import (
	"context"
	_ "effectiveMobile/docs"
	"effectiveMobile/internal/app/client/agify"
	"effectiveMobile/internal/app/client/genderize"
	"effectiveMobile/internal/app/client/nationalize"
	"effectiveMobile/internal/model"
	"effectiveMobile/internal/store"
	"encoding/json"
	"errors"
	"github.com/go-chi/chi/v5"
	httpSwagger "github.com/swaggo/http-swagger"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"golang.org/x/sync/errgroup"
	"net/http"
	"strconv"
	"time"
)

const (
	ErrJsonDecodeError        = "error decoding request"
	ErrNameAndSurnameRequired = "name and surname required"
	ErrInternalServer         = "internal server error"
	//ErrUnsupportedMediaType   = "unsupported media type"
)

type server struct {
	router      chi.Router
	config      *Config
	logger      *zap.Logger
	store       store.Store
	agify       *agify.Agify
	genderize   *genderize.Genderize
	nationalize *nationalize.Nationalize
}

func newServer(store store.Store, config *Config) *server {
	lvl, err := zapcore.ParseLevel(config.Zap.Level)
	if err != nil {
		panic(err)
	}
	logger, err := zap.NewDevelopment(zap.IncreaseLevel(lvl))

	if err != nil {
		panic(err)
	}
	return &server{
		router:      chi.NewRouter(),
		config:      config,
		logger:      logger,
		store:       store,
		agify:       agify.New(config.ExternalService.AgifyURL),
		genderize:   genderize.New(config.ExternalService.GenderizeURL),
		nationalize: nationalize.New(config.ExternalService.NationalizeURL),
	}
}

func (s *server) configureRouter() {
	s.router.Mount("/swagger", httpSwagger.WrapHandler)
	s.router.Route("/humans", func(r chi.Router) {
		r.Get("/", s.getHumans())
		r.Post("/", s.addHuman())
		r.Delete("/", s.deleteHuman())
		r.Patch("/", s.updateHuman())
	})
}

// addHuman adds a new human record
// @Summary Create a human
// @Description Create a new human with auto-filled age, gender, nationality
// @Tags humans
// @Accept application/json
// @Produce application/json
// @Param body body addHumanRequest true "Add Human payload"
// @Success 201 {object} model.Human
// @Failure 400 {object} string "name and surname required"
// @Failure 500 {string} string "Internal Server Error"
// @Router /humans [put]
func (s *server) addHuman() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if contentType := r.Header.Get("Content-Type"); contentType != "application/json" {
			http.Error(w, "Content-Type must be application/json", http.StatusUnsupportedMediaType)
			return
		}
		var req addHumanRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			s.logger.Info("error decoding request", zap.Error(err))
			http.Error(w, ErrJsonDecodeError, http.StatusBadRequest)
			return
		}
		if req.Name == "" || req.Surname == "" {
			http.Error(w, ErrNameAndSurnameRequired, http.StatusBadRequest)
			return
		}

		human := model.Human{
			Name:       req.Name,
			Surname:    req.Surname,
			Patronymic: req.Patronymic,
		}

		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		g, ctx := errgroup.WithContext(ctx)

		var (
			age         = 0
			gender      = "unknown"
			nationality = ""
		)

		// Запрос к Agify
		g.Go(func() error {
			resp, err := s.agify.Get(req.Name)
			if err != nil {
				s.logger.Error("agify Get Error", zap.Error(err))
				return err
			}
			select {
			case <-ctx.Done():
				s.logger.Warn("agify Get Timeout")
				return ctx.Err()
			default:
				age = resp.Age
				s.logger.Info("agify Get Success", zap.Any("resp", resp))
				return nil
			}
		})

		// Запрос к Genderize
		g.Go(func() error {
			resp, err := s.genderize.Get(req.Name)
			if err != nil {
				s.logger.Error("genderize Get Error", zap.Error(err))
				return err
			}
			select {
			case <-ctx.Done():
				s.logger.Info("genderize Get Timeout")
				return ctx.Err()
			default:
				if resp.Gender == "male" || resp.Gender == "female" {
					gender = resp.Gender
				}
				s.logger.Info("genderize Get Success", zap.Any("resp", resp))
				return nil
			}
		})

		// Запрос к Nationalize
		g.Go(func() error {
			resp, err := s.nationalize.Get(req.Name)
			if err != nil {
				s.logger.Error("nationalize Get Error", zap.Error(err))
				return err
			}
			select {
			case <-ctx.Done():
				s.logger.Info("nationalize Get Timeout")
				return ctx.Err()
			default:
				if len(resp.Country) > 0 {
					nationality = resp.Country[0].CountryId
				}
				s.logger.Info("nationalize Get Success", zap.Any("resp", resp))
				return nil
			}
		})

		// Ждём завершения всех горутин или таймаута
		if err := g.Wait(); err != nil {
			if errors.Is(err, context.DeadlineExceeded) {
				s.logger.Warn("one or more services timed out")
			} else {
				s.logger.Error("Failed to send request", zap.Error(err))
			}
		}

		human.Age = age
		human.Gender = gender
		human.Nationality = nationality

		s.logger.Info("added Human", zap.Any("human", human))

		if err := s.store.Human().AddHuman(r.Context(), &human); err != nil {
			s.logger.Error("failed to save human", zap.Error(err))
			http.Error(w, ErrJsonDecodeError, http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		if err := json.NewEncoder(w).Encode(human); err != nil {
			s.logger.Error("error encoding response", zap.Error(err))
		}
	}
}

// getHumans retrieves filtered list of humans
// @Summary Get humans
// @Description Retrieve humans with optional filtering and pagination
// @Tags humans
// @Accept json
// @Produce json
// @Param name query string false "Name filter"
// @Param surname query string false "Surname filter"
// @Param patronymic query string false "Patronymic filter"
// @Param gender query string false "Gender filter"
// @Param nationality query string false "Nationality filter"
// @Param min_age query int false "Minimum age filter"
// @Param max_age query int false "Maximum age filter"
// @Param page query int false "Page number"
// @Param page_size query int false "Page size"
// @Success 200 {array} model.Human
// @Failure 500 {string} string "Internal Server Error"
// @Router /humans [get]
func (s *server) getHumans() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		f := &model.HumanFilter{
			Name:        q.Get("name"),
			Surname:     q.Get("surname"),
			Patronymic:  q.Get("patronymic"),
			Gender:      q.Get("gender"),
			Nationality: q.Get("nationality"),
		}

		parseInt := func(key string, dest *int) {
			if s := q.Get(key); s != "" {
				if v, err := strconv.Atoi(s); err == nil {
					*dest = v
				}
			}
		}
		parseInt("id", &f.ID)
		parseInt("min_age", &f.MinAge)
		parseInt("max_age", &f.MaxAge)
		parseInt("page", &f.Page)
		parseInt("page_size", &f.PageSize)

		if f.Page < 1 {
			f.Page = 1
		}
		if f.PageSize <= 0 || f.PageSize > 100 {
			f.PageSize = 20
		}
		s.logger.Info("get humans", zap.Any("filter", f))

		humans, err := s.store.Human().GetHumans(r.Context(), f)
		if err != nil {
			s.logger.Error("failed to get humans", zap.Error(err))
			http.Error(w, ErrInternalServer, http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(humans); err != nil {
			s.logger.Error("error encoding response", zap.Error(err))
		}
		return
	}
}

// deleteHuman deletes a human by ID
// @Summary Delete human
// @Description Delete a human record by ID
// @Tags humans
// @Accept json
// @Param id body apiserver.deleteHumanRequest true "Delete Human request"
// @Success 200 {string} string "OK"
// @Failure 400 {string} string "Bad Request"
// @Failure 415 {string} string "Unsupported Media Type"
// @Failure 500 {string} string "Internal Server Error"
// @Router /humans [delete]
func (s *server) deleteHuman() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if contentType := r.Header.Get("Content-Type"); contentType != "application/json" {
			http.Error(w, "Content-Type must be application/json", http.StatusUnsupportedMediaType)
			return
		}
		req := deleteHumanRequest{}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, ErrJsonDecodeError, http.StatusBadRequest)
			return
		}
		if err := s.store.Human().DeleteHuman(r.Context(), req.ID); err != nil {
			s.logger.Error("failed to delete human", zap.Error(err))
			http.Error(w, ErrInternalServer, http.StatusInternalServerError)
			return
		}
		s.logger.Info("deleted human", zap.Int("id", req.ID))
		w.WriteHeader(http.StatusOK)
		return
	}
}

// updateHuman updates an existing human record
// @Summary Update human
// @Description Update human fields by ID
// @Tags humans
// @Accept json
// @Param human body apiserver.updateHumanRequest true "Update Human request"
// @Success 200 {string} string "OK"
// @Failure 400 {string} string "Bad Request"
// @Failure 415 {string} string "Unsupported Media Type"
// @Failure 500 {string} string "Internal Server Error"
// @Router /humans [patch]
func (s *server) updateHuman() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if contentType := r.Header.Get("Content-Type"); contentType != "application/json" {
			http.Error(w, "Content-Type must be application/json", http.StatusUnsupportedMediaType)
			return
		}
		req := updateHumanRequest{}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, ErrJsonDecodeError, http.StatusBadRequest)
			return
		}

		human := model.Human{
			Id:          req.ID,
			Name:        req.Name,
			Surname:     req.Surname,
			Patronymic:  req.Patronymic,
			Age:         req.Age,
			Gender:      req.Gender,
			Nationality: req.Nationality,
		}

		if err := s.store.Human().UpdateHuman(r.Context(), &human); err != nil {
			s.logger.Error("failed to update human", zap.Error(err))
			http.Error(w, ErrInternalServer, http.StatusInternalServerError)
			return
		}
		s.logger.Info("updated human", zap.Any("human", human))
		w.WriteHeader(http.StatusOK)
		return

	}
}
