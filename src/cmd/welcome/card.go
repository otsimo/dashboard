package main

import (
	"time"

	"bytes"
	"fmt"

	pb "github.com/otsimo/otsimopb"
	"github.com/Sirupsen/logrus"
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
	score := 250 + 250.0 * (float32(ttl) / float32(OneWeek))
	txt := createText(in.Language, profile)

	logrus.Debugf("card.go: ttl=%d OneWeek=%d score=%d", ttl, OneWeek, int32(score))

	return &pb.Card{
		Id:         NewUUID(),
		CreatedAt:  now,
		Text:       txt,
		ExpiresAt:  now + ttl,
		Decoration: &dec,
		Score:      int32(score),
		Data:       &pb.Card_Empty{Empty: &pb.CardEmpty{}},
	}
}
