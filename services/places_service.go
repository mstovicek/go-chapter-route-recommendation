package services

import (
	log "github.com/Sirupsen/logrus"
	"github.com/mstovicek/go-chapter-route-recommendation/entity"
	"sync"
)

const cachePrefixPlace = "place_"
const cachePrefixSuggestion = "suggestion_"

type placesService struct {
	cache Cache
	api   PlacesAPI
}

func NewPlacesService(c Cache, a PlacesAPI) *placesService {
	return &placesService{
		c,
		a,
	}
}

func (s *placesService) GetPlacesCollectionByPlaceIds(placeIds []string) []entity.Place {
	var waitGroup sync.WaitGroup

	placesChan := make(chan *entity.Place)

	placesCollection := []entity.Place{}

	for _, placeID := range placeIds {
		waitGroup.Add(1)
		go func(placeID string) {
			p := s.getPlace(placeID)
			placesChan <- p
		}(placeID)
	}

	go func() {
		for place := range placesChan {
			if place != nil {
				placesCollection = append(placesCollection, *place)
			}
			waitGroup.Done()
		}
	}()

	waitGroup.Wait()

	return placesCollection
}

func (s *placesService) getPlace(placeID string) *entity.Place {
	cachedPlace, found := s.cache.Get(cachePrefixPlace + placeID)
	if found {
		log.WithFields(log.Fields{
			"placeID": placeID,
		}).Info("Returning cached place")

		return cachedPlace.(*entity.Place)
	}

	p, err := s.api.GetPlaceDetail(placeID)
	if err != nil {
		log.WithFields(log.Fields{
			"placeID": placeID,
			"err":     err.Error(),
		}).Error("Cannot get place")
		return nil
	}

	s.cache.Set(cachePrefixPlace+placeID, p, 0)

	return p
}

func (s *placesService) GetPlacesSuggestionsByKeyword(keyword string) []entity.Suggestion {
	cachedSuggestions, found := s.cache.Get(cachePrefixSuggestion + keyword)
	if found {
		log.WithFields(log.Fields{
			"keyword": keyword,
		}).Info("Returning cached suggestions")

		return cachedSuggestions.([]entity.Suggestion)
	}

	suggestions, err := s.api.GetPlaceAutocompleteSuggestions(keyword)
	if err != nil {
		log.WithFields(log.Fields{
			"keyword": keyword,
			"err":     err.Error(),
		}).Error("Cannot get suggestions")
		return []entity.Suggestion{}
	}

	s.cache.Set(cachePrefixSuggestion+keyword, suggestions, 0)

	return suggestions
}

func (s *placesService) GetPlacesDistance(placesIDs []string) entity.DistanceMatrix {
	distanceMatrix, err := s.api.GetPlacesDistance(placesIDs)
	if err != nil {
		log.WithFields(log.Fields{
			"places": placesIDs,
			"err":    err.Error(),
		}).Error("Cannot get distances")
		return entity.NewDistanceMatrix()
	}
	return distanceMatrix
}
