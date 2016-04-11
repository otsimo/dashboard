package main

import (
	"time"

	pb "github.com/otsimo/otsimopb"
)

var dec = pb.CardDecoration{
	Size_:           pb.LARGE,
	BackgroundStyle: pb.EMPTY,
	ImageUrl:        "",
	LeftIcon:        "",
	RightIcon:       "",
}

func NewCard(in *pb.DashboardGetRequest, ttl int64) *pb.Card {
	now := time.Now().Unix()
	score := float32(500.0*ttl) / OneWeek
	return &pb.Card{
		Id:         NewUUID(),
		CreatedAt:  now,
		Text:       "",
		ExpiresAt:  now + ttl,
		Decoration: &dec,
		Score:      int32(score),
		Data:       &pb.Card_Empty{Empty: &pb.CardEmpty{}},
	}
}
