package geocoding

import (
	"context"
	"log"
)

// ServiceInterface define as operações de negócio para geocodificação
type ServiceInterface interface {
	// LocationService processa a geocodificação de origem e destino
	LocationService(c context.Context, data RequestLocationDTO) (ResponseLocationDTO, error)
}

// Service implementa a lógica de negócio de geocodificação enviando para o fallback se necessário
type Service struct {
	repository       RepositoryInterface
	googleMapsAPIKey string
}

// NewService cria uma nova instância do serviço de geocodificação com suporte a chave do Google Maps
func NewService(repository RepositoryInterface, googleMapsAPIKey string) *Service {
	return &Service{
		repository:       repository,
		googleMapsAPIKey: googleMapsAPIKey,
	}
}

// LocationService resolve as coordenadas e endereços para um par de CEPs (origem e destino)
func (s *Service) LocationService(ctx context.Context, data RequestLocationDTO) (ResponseLocationDTO, error) {
	// 1. Resolve Origem e Destino
	// 1.1 Resolve as informações do CEP de Origem (Consulta local -> Fallback Externo)
	originInfo, err := s.resolveCEP(ctx, data.Origin)
	if err != nil {
		log.Printf("[LocationService] Erro ao resolver CEP de origem [%s]: %v", data.Origin, err)
	}
	// 1.2 Resolve as informações do CEP de Destino (Consulta local -> Fallback Externo)
	destInfo, err := s.resolveCEP(ctx, data.Destination)
	if err != nil {
		log.Printf("[LocationService] Erro ao resolver CEP de destino [%s]: %v", data.Destination, err)
	}

	// 2. Busca informações da Viagem (Distância, Tempo, Polyline)
	var trips []TripInfo

	// OSRM (Motor 1)
	osrmTrips, errOSRM := s.fetchRouteInfo(ctx, originInfo, destInfo)
	if errOSRM != nil {
		log.Printf("[OSRM] Erro ao buscar rota: %v", errOSRM)
	} else {
		trips = append(trips, osrmTrips...)
	}

	// Valhalla (Motor 2)
	valhallaTrips, errVal := s.fetchValhallaRoute(ctx, originInfo, destInfo, data.Costing)
	if errVal != nil {
		log.Printf("[Valhalla] Erro ao buscar rota: %v", errVal)
	} else {
		for _, trip := range valhallaTrips {
			if trip.DistanceKm > 0 {
				trips = append(trips, trip)
			}
		}
	}

	// 4. Links de Navegação
	googleLink, wazeLink := s.generateNavigationLinks(originInfo, destInfo)

	return ResponseLocationDTO{
		Origin:      originInfo,
		Destination: destInfo,
		RoutesLink: RoutesLink{
			GoogleMapsLink: googleLink,
			WazeLink:       wazeLink,
		},
		TripInfo: trips,
	}, nil
}
