package advertisement

import (
	"errors"
	"time"
)

func (data *CreateAdvertisementRequest) ValidateCreate() error {
	if data.ExpirationDate.Before(data.PickupDate) {
		return errors.New("a data de expiração não pode ser anterior à data de retirada")
	}
	if data.DeliveryDate.Before(data.PickupDate) {
		return errors.New("a data de entrega não pode ser anterior à data de retirada")
	}
	now := time.Now()
	if data.PickupDate.Before(now) {
		return errors.New("a data de retirada não pode estar no passado")
	}

	if data.Title == "" {
		return errors.New("o título é obrigatório")
	}
	if data.Origin == "" || data.Destination == "" {
		return errors.New("a origem e o destino são obrigatórios")
	}
	if data.CargoSpecies == "" {
		return errors.New("a espécie da carga é obrigatória")
	}
	if data.Description == "" {
		return errors.New("a descrição é obrigatória")
	}
	if data.PaymentType == "" {
		return errors.New("o tipo de pagamento é obrigatório")
	}
	if data.Advance == "" {
		return errors.New("o campo de adiantamento é obrigatório")
	}
	if data.Situation == "" {
		return errors.New("a situação é obrigatória")
	}

	if data.Distance <= 0 {
		return errors.New("a distância deve ser maior que zero")
	}

	return nil
}

func (data *UpdateAdvertisementRequest) ValidateUpdate() error {
	if data.ExpirationDate.Before(data.PickupDate) {
		return errors.New("a data de expiração não pode ser anterior à data de retirada")
	}
	if data.DeliveryDate.Before(data.PickupDate) {
		return errors.New("a data de entrega não pode ser anterior à data de retirada")
	}
	now := time.Now()
	if data.PickupDate.Before(now) {
		return errors.New("a data de retirada não pode estar no passado")
	}

	if data.Title == "" {
		return errors.New("o título é obrigatório")
	}
	if data.Origin == "" || data.Destination == "" {
		return errors.New("a origem e o destino são obrigatórios")
	}
	if data.CargoType == "" {
		return errors.New("o tipo da carga é obrigatório")
	}
	if data.CargoSpecies == "" {
		return errors.New("a espécie da carga é obrigatória")
	}
	if data.Description == "" {
		return errors.New("a descrição é obrigatória")
	}
	if data.PaymentType == "" {
		return errors.New("o tipo de pagamento é obrigatório")
	}
	if data.Advance == "" {
		return errors.New("o campo de adiantamento é obrigatório")
	}
	if data.Situation == "" {
		return errors.New("a situação é obrigatória")
	}

	if data.Distance <= 0 {
		return errors.New("a distância deve ser maior que zero")
	}

	return nil
}
