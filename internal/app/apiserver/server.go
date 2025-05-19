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
	"github.com/go-chi/chi/v5"
	httpSwagger "github.com/swaggo/http-swagger"
	"go.uber.org/zap"
	"net/http"
	"strconv"
	"sync"
	"time"
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
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	return &server{
		router:      chi.NewRouter(),
		config:      config,
		logger:      logger,
		store:       store,
		agify:       agify.New(),
		genderize:   genderize.New(),
		nationalize: nationalize.New(),
	}
}

func (s *server) configureRouter() {
	s.router.Mount("/swagger", httpSwagger.WrapHandler)
	s.router.Put("/add_human", s.addHuman())
	s.router.Get("/get_humans", s.getHumans())
	s.router.Delete("/delete_human", s.deleteHuman())
	s.router.Patch("/update_human", s.updateHuman())
}

// addHuman adds a new human record
// @Summary Add human
// @Description Create a new human with auto-filled age, gender, nationality
// @Tags humans
// @Accept json
// @Produce json
// @Param human body apiserver.addHumanRequest true "Add Human request"
// @Success 201 {object} model.Human
// @Failure 400 {error} err.Error() Bad Request
// @Failure 415 {string} Content-Type must be application/json "Unsupported Media Type"
// @Failure 500 {string} Internal server error "Internal Server Error"
// @Router /add_human [put]
func (s *server) addHuman() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if contentType := r.Header.Get("Content-Type"); contentType != "application/json" {
			http.Error(w, "Content-Type must be application/json", http.StatusUnsupportedMediaType)
			return
		}
		req := addHumanRequest{}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if req.Name == "" || req.Surname == "" {
			http.Error(w, "Name and surname are required", http.StatusBadRequest)
			return
		}
		human := model.Human{
			Name:       req.Name,
			Surname:    req.Surname,
			Patronymic: req.Patronymic,
		}
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		ageChan := make(chan int, 1)
		genderChan := make(chan string, 1)
		nationalityChan := make(chan string, 1)

		defaultAge := 0
		defaultGender := "unknown"
		defaultNationality := ""

		var wg sync.WaitGroup
		wg.Add(3)

		go func() {
			defer wg.Done()
			resp, err := s.agify.Get(req.Name)
			if err != nil {
				s.logger.Error("agify Get Error", zap.Error(err))
				ageChan <- defaultAge
				return
			}
			select {
			case ageChan <- resp.Age:
			case <-ctx.Done():
				return
			}
		}()

		go func() {
			defer wg.Done()
			resp, err := s.genderize.Get(req.Name)
			if err != nil {
				s.logger.Error("genderize Get Error", zap.Error(err))
				genderChan <- defaultGender
				return
			}
			gender := defaultGender
			if resp.Gender == "male" {
				gender = "male"
			} else if resp.Gender == "female" {
				gender = "female"
			}
			select {
			case genderChan <- gender:
			case <-ctx.Done():
				return
			}
		}()

		// Nationality goroutine
		go func() {
			defer wg.Done()
			resp, err := s.nationalize.Get(req.Name)
			if err != nil {
				s.logger.Error("nationalize Get Error", zap.Error(err))
				nationalityChan <- defaultNationality
				return
			}
			nationality := defaultNationality
			if len(resp.Country) > 0 {
				nationality = resp.Country[0].CountryId
			}
			select {
			case nationalityChan <- nationality:
			case <-ctx.Done():
				return
			}
		}()

		go func() {
			wg.Wait()
			close(ageChan)
			close(genderChan)
			close(nationalityChan)
		}()

		select {
		case human.Age = <-ageChan:
		case <-ctx.Done():
			human.Age = defaultAge
		}

		select {
		case human.Gender = <-genderChan:
		case <-ctx.Done():
			human.Gender = defaultGender
		}

		select {
		case human.Nationality = <-nationalityChan:
		case <-ctx.Done():
			human.Nationality = defaultNationality
		}

		s.logger.Info("added Human", zap.Any("human", human))

		err := s.store.Human().AddHuman(r.Context(), &human)
		if err != nil {
			s.logger.Error("failed to save human", zap.Error(err))
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// Send response
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
// @Router /get_humans [get]
func (s *server) getHumans() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Парсим параметры
		q := r.URL.Query()
		f := &model.HumanFilter{
			Name:        q.Get("name"),
			Surname:     q.Get("surname"),
			Patronymic:  q.Get("patronymic"),
			Gender:      q.Get("gender"),
			Nationality: q.Get("nationality"),
		}

		// Вспомогательная функция для int-параметров
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

		// Установим разумные дефолты, если не переданы
		if f.Page < 1 {
			f.Page = 1
		}
		if f.PageSize <= 0 || f.PageSize > 100 {
			f.PageSize = 20
		}

		humans, err := s.store.Human().GetHumans(r.Context(), f)
		if err != nil {
			s.logger.Error("failed to get humans", zap.Error(err))
			http.Error(w, "Internal server error", http.StatusInternalServerError)
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
// @Router /delete_human [delete]
func (s *server) deleteHuman() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if contentType := r.Header.Get("Content-Type"); contentType != "application/json" {
			http.Error(w, "Content-Type must be application/json", http.StatusUnsupportedMediaType)
			return
		}
		req := deleteHumanRequest{}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err := s.store.Human().DeleteHuman(r.Context(), req.ID); err != nil {
			s.logger.Error("failed to delete human", zap.Error(err))
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
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
// @Router /update_human [patch]
func (s *server) updateHuman() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if contentType := r.Header.Get("Content-Type"); contentType != "application/json" {
			http.Error(w, "Content-Type must be application/json", http.StatusUnsupportedMediaType)
			return
		}
		req := updateHumanRequest{}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
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
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		return

	}
}
