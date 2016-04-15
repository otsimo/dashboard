package main

import (
	"bytes"
	"fmt"
	"time"

	pb "github.com/otsimo/otsimopb"
)

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

func NewCard(in *pb.DashboardGetRequest, ttl int64, profile *pb.Profile, id int64) *pb.Card {
	now := time.Now().Unix()
	txt := createText(in.Language, profile)
	cid := fmt.Sprintf("welcome-%s-%s-%d", in.ProfileId, in.Language, id)
	return &pb.Card{
		Id:            cid,
		CreatedAt:     now,
		Text:          txt,
		Language:      in.Language,
		ExpiresAt:     now + ttl,
		Decoration:    &pb.CardDecoration{
			Size_:           pb.LARGE,
			BackgroundStyle: pb.EMPTY,
			ImageUrl:        "",
			LeftIcon:        "time",
			RightIcon:       "",
		},
		Data:          &pb.Card_Empty{Empty: &pb.CardEmpty{}},
	}
}
