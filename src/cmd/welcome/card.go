package main

import (
	"time"

	"bytes"
	"fmt"

	"github.com/Sirupsen/logrus"
	pb "github.com/otsimo/otsimopb"
)

var dec = pb.CardDecoration{
	Size_:           pb.LARGE,
	BackgroundStyle: pb.EMPTY,
	ImageUrl:        "",
	LeftIcon:        "",
	RightIcon:       "",
}

func createText(lang string, profile *pb.Profile) string {
	var b bytes.Buffer
	data := map[string]interface{}{
		"FirstName": profile.FirstName,
		"LastName":  profile.LastName,
		"Email":     profile.Email,
	}
	templates.ExecuteTemplate(&b, fmt.Sprintf("%s.tmpl", lang), data)
	return b.String()
}

func NewCard(in *pb.DashboardGetRequest, ttl int64, profile *pb.Profile) *pb.Card {
	now := time.Now().Unix()
	score := 500
	txt := createText(in.Language, profile)

	return &pb.Card{
		Id:            NewUUID(),
		CreatedAt:     now,
		Text:          txt,
		ExpiresAt:     now + ttl,
		Decoration:    &dec,
		ProviderScore: int32(score),
		Data:          &pb.Card_Empty{Empty: &pb.CardEmpty{}},
	}
}
