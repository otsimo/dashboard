package main

import (
	"bytes"
	"fmt"
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

	id := fmt.Sprintf("%s-%s", in.ProfileId, in.Language)

	return &pb.Card{
		Id:            id,
		CreatedAt:     now,
		Text:          txt,
		Language:      in.Language,
		ExpiresAt:     now + ttl,
		Decoration:    &dec,
		ProviderScore: int32(score),
		Data:          &pb.Card_Empty{Empty: &pb.CardEmpty{}},
	}
}
