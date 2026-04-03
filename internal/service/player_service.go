package service

import (
	"context"
	"time"

	"bidxi/internal/models"
	"bidxi/internal/repository"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type PlayerService interface {
	SeedInitialPlayers(ctx context.Context) error
}

type playerService struct {
	playerRepo repository.PlayerRepository
}

func NewPlayerService(playerRepo repository.PlayerRepository) PlayerService {
	return &playerService{playerRepo: playerRepo}
}

func (s *playerService) SeedInitialPlayers(ctx context.Context) error {
	// 1. Check if we already have players to avoid wiping IDs on every restart
	existing, _ := s.playerRepo.GetAll(ctx)
	if len(existing) > 0 {
		return nil
	}

	// 2. Wipe only if empty, though technically not needed if we return early above
	s.playerRepo.ClearAll(ctx)

	players := []models.Player{
		// --- MARQUEE ---
		{Name: "Virat Kohli", Role: "BAT", Nationality: "Indian", Category: "MARQUEE", BasePrice: 200, ImageUrl: "https://documents.iplt20.com/ipl/IPLHeadshot2024/2.png"},
		{Name: "MS Dhoni", Role: "WK-BAT", Nationality: "Indian", Category: "MARQUEE", BasePrice: 200, ImageUrl: "https://documents.iplt20.com/ipl/IPLHeadshot2024/57.png"},
		{Name: "Rohit Sharma", Role: "BAT", Nationality: "Indian", Category: "MARQUEE", BasePrice: 200, ImageUrl: "https://documents.iplt20.com/ipl/IPLHeadshot2024/6.png"},
		{Name: "Hardik Pandya", Role: "ALL", Nationality: "Indian", Category: "MARQUEE", BasePrice: 200, ImageUrl: "https://documents.iplt20.com/ipl/IPLHeadshot2024/54.png"},
		{Name: "Pat Cummins", Role: "BOWL", Nationality: "Overseas", Category: "MARQUEE", BasePrice: 200, ImageUrl: "https://documents.iplt20.com/ipl/IPLHeadshot2024/446.png"},
		{Name: "Travis Head", Role: "BAT", Nationality: "Overseas", Category: "MARQUEE", BasePrice: 200, ImageUrl: "https://documents.iplt20.com/ipl/IPLHeadshot2024/1020.png"},
		{Name: "Jasprit Bumrah", Role: "BOWL", Nationality: "Indian", Category: "MARQUEE", BasePrice: 200, ImageUrl: "https://documents.iplt20.com/ipl/IPLHeadshot2024/9.png"},
		{Name: "Rashid Khan", Role: "BOWL", Nationality: "Overseas", Category: "MARQUEE", BasePrice: 200, ImageUrl: "https://documents.iplt20.com/ipl/IPLHeadshot2024/218.png"},
		{Name: "Shubman Gill", Role: "BAT", Nationality: "Indian", Category: "MARQUEE", BasePrice: 200, ImageUrl: "https://documents.iplt20.com/ipl/IPLHeadshot2024/62.png"},
		{Name: "Rishabh Pant", Role: "WK-BAT", Nationality: "Indian", Category: "MARQUEE", BasePrice: 200, ImageUrl: "https://documents.iplt20.com/ipl/IPLHeadshot2024/31.png"},

		// --- WICKETKEEPERS ---
		{Name: "Sanju Samson", Role: "WK-BAT", Nationality: "Indian", Category: "WK", BasePrice: 150, ImageUrl: "https://documents.iplt20.com/ipl/IPLHeadshot2024/190.png"},
		{Name: "KL Rahul", Role: "WK-BAT", Nationality: "Indian", Category: "WK", BasePrice: 150, ImageUrl: "https://documents.iplt20.com/ipl/IPLHeadshot2024/19.png"},
		{Name: "Ishan Kishan", Role: "WK-BAT", Nationality: "Indian", Category: "WK", BasePrice: 150, ImageUrl: "https://documents.iplt20.com/ipl/IPLHeadshot2024/164.png"},
		{Name: "Heinrich Klaasen", Role: "WK-BAT", Nationality: "Overseas", Category: "WK", BasePrice: 200, ImageUrl: "https://documents.iplt20.com/ipl/IPLHeadshot2024/202.png"},
		{Name: "Quinton de Kock", Role: "WK-BAT", Nationality: "Overseas", Category: "WK", BasePrice: 150, ImageUrl: "https://documents.iplt20.com/ipl/IPLHeadshot2024/170.png"},
		{Name: "Nicholas Pooran", Role: "WK-BAT", Nationality: "Overseas", Category: "WK", BasePrice: 150, ImageUrl: "https://documents.iplt20.com/ipl/IPLHeadshot2024/173.png"},
		{Name: "Dhruv Jurel", Role: "WK-BAT", Nationality: "Indian", Category: "WK", BasePrice: 50, ImageUrl: "https://documents.iplt20.com/ipl/IPLHeadshot2024/2033.png"},

		// --- BATSMEN ---
		{Name: "Suryakumar Yadav", Role: "BAT", Nationality: "Indian", Category: "BATSMAN", BasePrice: 150, ImageUrl: "https://documents.iplt20.com/ipl/IPLHeadshot2024/174.png"},
		{Name: "Ruturaj Gaikwad", Role: "BAT", Nationality: "Indian", Category: "BATSMAN", BasePrice: 150, ImageUrl: "https://documents.iplt20.com/ipl/IPLHeadshot2024/102.png"},
		{Name: "Yashasvi Jaiswal", Role: "BAT", Nationality: "Indian", Category: "BATSMAN", BasePrice: 150, ImageUrl: "https://documents.iplt20.com/ipl/IPLHeadshot2024/533.png"},
		{Name: "Rinku Singh", Role: "BAT", Nationality: "Indian", Category: "BATSMAN", BasePrice: 80, ImageUrl: "https://documents.iplt20.com/ipl/IPLHeadshot2024/152.png"},
		{Name: "David Warner", Role: "BAT", Nationality: "Overseas", Category: "BATSMAN", BasePrice: 150, ImageUrl: "https://documents.iplt20.com/ipl/IPLHeadshot2024/7.png"},
		{Name: "Faf du Plessis", Role: "BAT", Nationality: "Overseas", Category: "BATSMAN", BasePrice: 150, ImageUrl: "https://documents.iplt20.com/ipl/IPLHeadshot2024/24.png"},
		{Name: "Tilak Varma", Role: "BAT", Nationality: "Indian", Category: "BATSMAN", BasePrice: 50, ImageUrl: "https://documents.iplt20.com/ipl/IPLHeadshot2024/540.png"},

		// --- ALLROUNDERS ---
		{Name: "Ravindra Jadeja", Role: "ALL", Nationality: "Indian", Category: "ALLROUNDER", BasePrice: 200, ImageUrl: "https://documents.iplt20.com/ipl/IPLHeadshot2024/46.png"},
		{Name: "Glenn Maxwell", Role: "ALL", Nationality: "Overseas", Category: "ALLROUNDER", BasePrice: 200, ImageUrl: "https://documents.iplt20.com/ipl/IPLHeadshot2024/15.png"},
		{Name: "Andre Russell", Role: "ALL", Nationality: "Overseas", Category: "ALLROUNDER", BasePrice: 200, ImageUrl: "https://documents.iplt20.com/ipl/IPLHeadshot2024/141.png"},
		{Name: "Sam Curran", Role: "ALL", Nationality: "Overseas", Category: "ALLROUNDER", BasePrice: 200, ImageUrl: "https://documents.iplt20.com/ipl/IPLHeadshot2024/182.png"},
		{Name: "Venkatesh Iyer", Role: "ALL", Nationality: "Indian", Category: "ALLROUNDER", BasePrice: 100, ImageUrl: "https://documents.iplt20.com/ipl/IPLHeadshot2024/541.png"},
		{Name: "Rachin Ravindra", Role: "ALL", Nationality: "Overseas", Category: "ALLROUNDER", BasePrice: 150, ImageUrl: "https://documents.iplt20.com/ipl/IPLHeadshot2024/1021.png"},

		// --- BOWLERS ---
		{Name: "Mitchell Starc", Role: "BOWL", Nationality: "Overseas", Category: "BOWLER", BasePrice: 200, ImageUrl: "https://documents.iplt20.com/ipl/IPLHeadshot2024/55.png"},
		{Name: "Mohammed Shami", Role: "BOWL", Nationality: "Indian", Category: "BOWLER", BasePrice: 200, ImageUrl: "https://documents.iplt20.com/ipl/IPLHeadshot2024/47.png"},
		{Name: "Trent Boult", Role: "BOWL", Nationality: "Overseas", Category: "BOWLER", BasePrice: 200, ImageUrl: "https://documents.iplt20.com/ipl/IPLHeadshot2024/66.png"},
		{Name: "Kuldeep Yadav", Role: "BOWL", Nationality: "Indian", Category: "BOWLER", BasePrice: 150, ImageUrl: "https://documents.iplt20.com/ipl/IPLHeadshot2024/151.png"},
		{Name: "Yuzvendra Chahal", Role: "BOWL", Nationality: "Indian", Category: "BOWLER", BasePrice: 150, ImageUrl: "https://documents.iplt20.com/ipl/IPLHeadshot2024/111.png"},
		{Name: "Mohammed Siraj", Role: "BOWL", Nationality: "Indian", Category: "BOWLER", BasePrice: 150, ImageUrl: "https://documents.iplt20.com/ipl/IPLHeadshot2024/63.png"},
		{Name: "Arshdeep Singh", Role: "BOWL", Nationality: "Indian", Category: "BOWLER", BasePrice: 150, ImageUrl: "https://documents.iplt20.com/ipl/IPLHeadshot2024/535.png"},
		{Name: "Sunil Narine", Role: "BOWL", Nationality: "Overseas", Category: "BOWLER", BasePrice: 200, ImageUrl: "https://documents.iplt20.com/ipl/IPLHeadshot2024/156.png"},
		{Name: "Jitesh Sharma", Role: "WK-BAT", Nationality: "Indian", Category: "WK", BasePrice: 50, ImageUrl: "https://documents.iplt20.com/ipl/IPLHeadshot2024/1000.png"},
		{Name: "Prabhsimran Singh", Role: "WK-BAT", Nationality: "Indian", Category: "WK", BasePrice: 20, ImageUrl: "https://documents.iplt20.com/ipl/IPLHeadshot2024/543.png"},
		{Name: "Jonny Bairstow", Role: "WK-BAT", Nationality: "Overseas", Category: "WK", BasePrice: 150, ImageUrl: "https://documents.iplt20.com/ipl/IPLHeadshot2024/506.png"},
	}

	for _, p := range players {
		p.ID = primitive.NewObjectID()
		p.CreatedAt = time.Now()
		if err := s.playerRepo.Create(ctx, &p); err != nil {
			return err
		}
	}

	return nil
}
