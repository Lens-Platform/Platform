package proto

import (
	"errors"

	"github.com/BlackspaceInc/BlackspacePlatform/src/services/shopper_service/pkg/graphql_api/model"
)

// ConvertToModel converts an object of type shopper account input to shopper account model
func (account *ShopperAccountInput) ConvertToModel() (*model.ShopperAccount, error) {
	if account == nil {
		return nil, errors.New("cannot convert an empty input shopper account")
	}

	addresses := account.ConvertAddresses()
	tags := account.ConvertTags(account.Tags)
	creditCard := account.ConvertCreditCard()
	subscribedTopics := account.ConvertAccountSubscribedTopics(account.SubscribedTopics)
	causes := account.ConvertAccountSubscribedTopics(account.Causes)

	return &model.ShopperAccount{
		Id:                  0,
		FirstName:           *account.FirstName,
		LastName:            *account.LastName,
		Email:               *account.Email,
		Username:            *account.Username,
		Phone:               *account.Phone,
		IsActive:            *account.IsActive,
		AcceptsMarketing:    *account.AcceptsMarketing,
		AcceptedMarketingAt: *account.AcceptedMarketingAt,
		Addresses:           addresses,
		Tags:                tags,
		Causes:              causes,
		CreditCard:          creditCard,
		SubscribedTopics:    subscribedTopics,
		AuthnId:             0,
		Password:            *account.Password,
	}, nil

}

// ConvertAccountSubscribedTopics convert subscribed topics
func (account *ShopperAccountInput) ConvertAccountSubscribedTopics(topics []*SubscribedTopicInput) []*model.SubscribedTopic {
	subscribedTopics := make([]*model.SubscribedTopic, 0)
	for _, topic := range topics {
		tags := account.ConvertTopicTags(topic)
		coverImage := account.ConvertImageData(topic)

		subscribedTopics = append(subscribedTopics, &model.SubscribedTopic{
			Id: 0,
			SubscribedTopic: &model.Topic{
				Id:              0,
				Name:            *topic.SubscribedTopic.Name,
				Tags:            tags,
				TopicCoverImage: coverImage,
			},
			SubscribedAt: *topic.SubscribedAt,
			Description:  *topic.Description,
		})
	}
	return subscribedTopics
}

// ConvertImageData converts image data
func (account *ShopperAccountInput) ConvertImageData(topic *SubscribedTopicInput) *model.Image {
	coverImg := topic.SubscribedTopic.TopicCoverImage
	return &model.Image{
		Id: 0,
		Metadata: &model.ImageMeta{
			Width:     int32(*coverImg.Metadata.Width),
			Height:    int32(*coverImg.Metadata.Height),
			RedData:   []byte(*coverImg.Metadata.RedData),
			GreenData: []byte(*coverImg.Metadata.GreenData),
			BlueData:  []byte(*coverImg.Metadata.BlueData),
		},
		BlobUrl: *coverImg.BlobURL,
		AltText: *coverImg.AltText,
	}
}

// ConvertTopicTags converts topic tags
func (account *ShopperAccountInput) ConvertTopicTags(topic *SubscribedTopicInput) []string {
	return account.ConvertTags(topic.SubscribedTopic.Tags)
}

// ConvertCreditCard converts credit card
func (account *ShopperAccountInput) ConvertCreditCard() *model.CreditCard {
	card := account.CreditCard
	creditCard := &model.CreditCard{
		Id:                       0,
		CardNumber:               *card.CardNumber,
		CardBrand:                *card.CardBrand,
		ExpiresSoon:              *card.ExpiresSoon,
		ExpirationMonth:          account.ConvertExpirationMonth(),
		ExpirationYear:           account.ConvertExpirationYear(),
		FirstDigits:              *card.FirstDigits,
		LastDigits:               *card.LastDigits,
		MaskedNumber:             *card.MaskedNumber,
		CardHolderName:           *card.CardHolderName,
		CreditCardBillingAddress: account.ConvertSingularAddress(account.CreditCard.CreditCardBillingAddress),
	}
	return creditCard
}

// ConvertExpirationYear converts the expiration year of a credit card
func (account *ShopperAccountInput) ConvertExpirationYear() int64 {
	return int64(*account.CreditCard.ExpirationYear)
}

// ConvertExpirationMonth converts the expiration month of a credit card
func (account *ShopperAccountInput) ConvertExpirationMonth() int64 {
	return int64(*account.CreditCard.ExpirationMonth)
}

// ConvertTags converts a set of account tag
func (account *ShopperAccountInput) ConvertTags(inputTags []*string) []string {
	tags := make([]string, 0)
	for _, tag := range inputTags {
		tags = append(tags, *tag)
	}
	return tags
}

// ConvertAddresses converts numerous addresses
func (account *ShopperAccountInput) ConvertAddresses() []*model.Address {
	var addresses = make([]*model.Address, 0)
	for _, address := range account.Addresses {
		addresses = append(addresses, account.ConvertSingularAddress(address))
	}
	return addresses
}

// ConvertSingularAddress converts a singular address
func (account *ShopperAccountInput) ConvertSingularAddress(address *AddressInput) *model.Address {
	return &model.Address{
		Id:           0,
		Street:       *address.Street,
		Province:     *address.Province,
		City:         *address.City,
		ZipCode:      *address.ZipCode,
		Country:      *address.Country,
		CountryCode:  *address.CountryCode,
		ProvinceCode: *address.ProvinceCode,
	}
}
