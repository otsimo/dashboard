package storage

import (
	"time"

	"github.com/otsimo/otsimopb"
)

type Item struct {
	ID           string    `bson:"_id" gorm:"primary_key"`
	UserID       string    `bson:"user_id"`
	ProviderName string    `bson:"provider_name"`
	CreatedAt    time.Time `bson:"created_at"`
	ExpiresAt    time.Time `bson:"expires_at"`
	Language     string    `bson:"language"`
	Card         []byte    `bson:"card"`
}

type ProviderUserInfo struct {
	Name      string    `bson:"name"`
	UserID    string    `bson:"-"`
	FetchedAt time.Time `bson:"fetched_at"`
	ExpiresAt time.Time `bson:"expires_at"`
}

type DashboardUser struct {
	ID        string             `bson:"_id" gorm:"type:varchar(36);primary_key"`
	Providers []ProviderUserInfo `bson:"providers" gorm:"ForeignKey:UserID"`
}

func (i *Item) GetCard() *otsimopb.Card {
	card := &otsimopb.Card{}
	if err := card.Unmarshal(i.Card); err != nil {
		return nil
	}
	return card
}
