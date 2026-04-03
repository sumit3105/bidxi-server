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
	// 1. Wipe old players to fix invalid ID references
	s.playerRepo.ClearAll(ctx)

	players := []models.Player{
		// --- MARQUEE (Top Tier) ---
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

		// --- WICKETKEEPERS (WK) ---
		{Name: "Sanju Samson", Role: "WK-BAT", Nationality: "Indian", Category: "WK", BasePrice: 150, ImageUrl: "https://documents.iplt20.com/ipl/IPLHeadshot2024/190.png"},
		{Name: "KL Rahul", Role: "WK-BAT", Nationality: "Indian", Category: "WK", BasePrice: 150, ImageUrl: "https://documents.iplt20.com/ipl/IPLHeadshot2024/19.png"},
		{Name: "Ishan Kishan", Role: "WK-BAT", Nationality: "Indian", Category: "WK", BasePrice: 150, ImageUrl: "https://documents.iplt20.com/ipl/IPLHeadshot2024/164.png"},
		{Name: "Heinrich Klaasen", Role: "WK-BAT", Nationality: "Overseas", Category: "WK", BasePrice: 200, ImageUrl: "https://documents.iplt20.com/ipl/IPLHeadshot2024/202.png"},
		{Name: "Quinton de Kock", Role: "WK-BAT", Nationality: "Overseas", Category: "WK", BasePrice: 150, ImageUrl: "https://documents.iplt20.com/ipl/IPLHeadshot2024/170.png"},
		{Name: "Phil Salt", Role: "WK-BAT", Nationality: "Overseas", Category: "WK", BasePrice: 150, ImageUrl: "https://documents.iplt20.com/ipl/IPLHeadshot2024/1018.png"},
		{Name: "Nicholas Pooran", Role: "WK-BAT", Nationality: "Overseas", Category: "WK", BasePrice: 150, ImageUrl: "https://documents.iplt20.com/ipl/IPLHeadshot2024/173.png"},

		// --- BATSMEN (BATSMAN) ---
		{Name: "Suryakumar Yadav", Role: "BAT", Nationality: "Indian", Category: "BATSMAN", BasePrice: 150, ImageUrl: "https://documents.iplt20.com/ipl/IPLHeadshot2024/174.png"},
		{Name: "Ruturaj Gaikwad", Role: "BAT", Nationality: "Indian", Category: "BATSMAN", BasePrice: 150, ImageUrl: "https://documents.iplt20.com/ipl/IPLHeadshot2024/102.png"},
		{Name: "Yashasvi Jaiswal", Role: "BAT", Nationality: "Indian", Category: "BATSMAN", BasePrice: 150, ImageUrl: "https://documents.iplt20.com/ipl/IPLHeadshot2024/533.png"},
		{Name: "Rinku Singh", Role: "BAT", Nationality: "Indian", Category: "BATSMAN", BasePrice: 80, ImageUrl: "https://documents.iplt20.com/ipl/IPLHeadshot2024/152.png"},
		{Name: "Jake Fraser-McGurk", Role: "BAT", Nationality: "Overseas", Category: "BATSMAN", BasePrice: 50, ImageUrl: "https://documents.iplt20.com/ipl/IPLHeadshot2024/1017.png"},
		{Name: "David Warner", Role: "BAT", Nationality: "Overseas", Category: "BATSMAN", BasePrice: 150, ImageUrl: "https://documents.iplt20.com/ipl/IPLHeadshot2024/7.png"},
		{Name: "Faf du Plessis", Role: "BAT", Nationality: "Overseas", Category: "BATSMAN", BasePrice: 150, ImageUrl: "https://documents.iplt20.com/ipl/IPLHeadshot2024/24.png"},
		{Name: "Kane Williamson", Role: "BAT", Nationality: "Overseas", Category: "BATSMAN", BasePrice: 150, ImageUrl: "https://documents.iplt20.com/ipl/IPLHeadshot2024/65.png"},
		{Name: "Sai Sudharsan", Role: "BAT", Nationality: "Indian", Category: "BATSMAN", BasePrice: 50, ImageUrl: "https://documents.iplt20.com/ipl/IPLHeadshot2024/521.png"},

		// --- ALLROUNDERS (ALLROUNDER) ---
		{Name: "Ravindra Jadeja", Role: "ALL", Nationality: "Indian", Category: "ALLROUNDER", BasePrice: 200, ImageUrl: "https://documents.iplt20.com/ipl/IPLHeadshot2024/46.png"},
		{Name: "Glenn Maxwell", Role: "ALL", Nationality: "Overseas", Category: "ALLROUNDER", BasePrice: 200, ImageUrl: "https://documents.iplt20.com/ipl/IPLHeadshot2024/15.png"},
		{Name: "Andre Russell", Role: "ALL", Nationality: "Overseas", Category: "ALLROUNDER", BasePrice: 200, ImageUrl: "https://documents.iplt20.com/ipl/IPLHeadshot2024/141.png"},
		{Name: "Marcus Stoinis", Role: "ALL", Nationality: "Overseas", Category: "ALLROUNDER", BasePrice: 200, ImageUrl: "https://documents.iplt20.com/ipl/IPLHeadshot2024/23.png"},
		{Name: "Axar Patel", Role: "ALL", Nationality: "Indian", Category: "ALLROUNDER", BasePrice: 150, ImageUrl: "https://documents.iplt20.com/ipl/IPLHeadshot2024/110.png"},
		{Name: "Liam Livingstone", Role: "ALL", Nationality: "Overseas", Category: "ALLROUNDER", BasePrice: 150, ImageUrl: "https://documents.iplt20.com/ipl/IPLHeadshot2024/181.png"},
		{Name: "Sam Curran", Role: "ALL", Nationality: "Overseas", Category: "ALLROUNDER", BasePrice: 200, ImageUrl: "https://documents.iplt20.com/ipl/IPLHeadshot2024/182.png"},
		{Name: "Cameron Green", Role: "ALL", Nationality: "Overseas", Category: "ALLROUNDER", BasePrice: 200, ImageUrl: "https://documents.iplt20.com/ipl/IPLHeadshot2024/550.png"},
		{Name: "Venkatesh Iyer", Role: "ALL", Nationality: "Indian", Category: "ALLROUNDER", BasePrice: 100, ImageUrl: "https://documents.iplt20.com/ipl/IPLHeadshot2024/541.png"},

		// --- BOWLERS (BOWLER) ---
		{Name: "Mitchell Starc", Role: "BOWL", Nationality: "Overseas", Category: "BOWLER", BasePrice: 200, ImageUrl: "https://documents.iplt20.com/ipl/IPLHeadshot2024/55.png"},
		{Name: "Mohammed Shami", Role: "BOWL", Nationality: "Indian", Category: "BOWLER", BasePrice: 200, ImageUrl: "https://documents.iplt20.com/ipl/IPLHeadshot2024/47.png"},
		{Name: "Trent Boult", Role: "BOWL", Nationality: "Overseas", Category: "BOWLER", BasePrice: 200, ImageUrl: "https://documents.iplt20.com/ipl/IPLHeadshot2024/66.png"},
		{Name: "Kuldeep Yadav", Role: "BOWL", Nationality: "Indian", Category: "BOWLER", BasePrice: 150, ImageUrl: "https://documents.iplt20.com/ipl/IPLHeadshot2024/151.png"},
		{Name: "Yuzvendra Chahal", Role: "BOWL", Nationality: "Indian", Category: "BOWLER", BasePrice: 150, ImageUrl: "https://documents.iplt20.com/ipl/IPLHeadshot2024/111.png"},
		{Name: "Matheesha Pathirana", Role: "BOWL", Nationality: "Overseas", Category: "BOWLER", BasePrice: 100, ImageUrl: "https://documents.iplt20.com/ipl/IPLHeadshot2024/1014.png"},
		{Name: "Mohammed Siraj", Role: "BOWL", Nationality: "Indian", Category: "BOWLER", BasePrice: 150, ImageUrl: "https://documents.iplt20.com/ipl/IPLHeadshot2024/63.png"},
		{Name: "Kagiso Rabada", Role: "BOWL", Nationality: "Overseas", Category: "BOWLER", BasePrice: 200, ImageUrl: "https://documents.iplt20.com/ipl/IPLHeadshot2024/116.png"},
		{Name: "Arshdeep Singh", Role: "BOWL", Nationality: "Indian", Category: "BOWLER", BasePrice: 150, ImageUrl: "https://documents.iplt20.com/ipl/IPLHeadshot2024/535.png"},
		{Name: "Nortje", Role: "BOWL", Nationality: "Overseas", Category: "BOWLER", BasePrice: 150, ImageUrl: "https://documents.iplt20.com/ipl/IPLHeadshot2024/166.png"},
		{Name: "Varun Chakaravarthy", Role: "BOWL", Nationality: "Indian", Category: "BOWLER", BasePrice: 100, ImageUrl: "https://documents.iplt20.com/ipl/IPLHeadshot2024/140.png"},
		{Name: "Sunil Narine", Role: "BOWL", Nationality: "Overseas", Category: "BOWLER", BasePrice: 200, ImageUrl: "https://documents.iplt20.com/ipl/IPLHeadshot2024/156.png"},
		// --- EXTRA GENERATED PLAYERS (for a long auction) ---
		{Name: "Rajat Patidar", Role: "BAT", Nationality: "Indian", Category: "BATSMAN", BasePrice: 50, ImageUrl: "https://documents.iplt20.com/ipl/IPLHeadshot2024/54.png"},
		{Name: "Shivam Dube", Role: "ALL", Nationality: "Indian", Category: "ALLROUNDER", BasePrice: 100, ImageUrl: "https://documents.iplt20.com/ipl/IPLHeadshot2024/54.png"},
		{Name: "Ajinkya Rahane", Role: "BAT", Nationality: "Indian", Category: "BATSMAN", BasePrice: 50, ImageUrl: "https://documents.iplt20.com/ipl/IPLHeadshot2024/54.png"},
		{Name: "Shardul Thakur", Role: "ALL", Nationality: "Indian", Category: "ALLROUNDER", BasePrice: 100, ImageUrl: "https://documents.iplt20.com/ipl/IPLHeadshot2024/54.png"},
		{Name: "Deepak Chahar", Role: "BOWL", Nationality: "Indian", Category: "BOWLER", BasePrice: 150, ImageUrl: "https://documents.iplt20.com/ipl/IPLHeadshot2024/54.png"},
		{Name: "Moeen Ali", Role: "ALL", Nationality: "Overseas", Category: "ALLROUNDER", BasePrice: 200, ImageUrl: "https://documents.iplt20.com/ipl/IPLHeadshot2024/54.png"},
		{Name: "Devon Conway", Role: "WK-BAT", Nationality: "Overseas", Category: "WK", BasePrice: 200, ImageUrl: "https://documents.iplt20.com/ipl/IPLHeadshot2024/54.png"},
		{Name: "Rachin Ravindra", Role: "ALL", Nationality: "Overseas", Category: "ALLROUNDER", BasePrice: 150, ImageUrl: "https://documents.iplt20.com/ipl/IPLHeadshot2024/54.png"},
		{Name: "Daryl Mitchell", Role: "BAT", Nationality: "Overseas", Category: "ALLROUNDER", BasePrice: 150, ImageUrl: "https://documents.iplt20.com/ipl/IPLHeadshot2024/54.png"},
		{Name: "Maheesh Theekshana", Role: "BOWL", Nationality: "Overseas", Category: "BOWLER", BasePrice: 100, ImageUrl: "https://documents.iplt20.com/ipl/IPLHeadshot2024/54.png"},
		{Name: "Ravi Bishnoi", Role: "BOWL", Nationality: "Indian", Category: "BOWLER", BasePrice: 100, ImageUrl: "https://documents.iplt20.com/ipl/IPLHeadshot2024/54.png"},
		{Name: "Krunal Pandya", Role: "ALL", Nationality: "Indian", Category: "ALLROUNDER", BasePrice: 100, ImageUrl: "https://documents.iplt20.com/ipl/IPLHeadshot2024/54.png"},
		{Name: "Ayush Badoni", Role: "BAT", Nationality: "Indian", Category: "BATSMAN", BasePrice: 20, ImageUrl: "https://documents.iplt20.com/ipl/IPLHeadshot2024/54.png"},
		{Name: "Deepak Hooda", Role: "BAT", Nationality: "Indian", Category: "BATSMAN", BasePrice: 50, ImageUrl: "https://documents.iplt20.com/ipl/IPLHeadshot2024/54.png"},
		{Name: "Avesh Khan", Role: "BOWL", Nationality: "Indian", Category: "BOWLER", BasePrice: 50, ImageUrl: "https://documents.iplt20.com/ipl/IPLHeadshot2024/54.png"},
		{Name: "Washington Sundar", Role: "ALL", Nationality: "Indian", Category: "ALLROUNDER", BasePrice: 50, ImageUrl: "https://documents.iplt20.com/ipl/IPLHeadshot2024/54.png"},
		{Name: "Abhishek Sharma", Role: "ALL", Nationality: "Indian", Category: "ALLROUNDER", BasePrice: 50, ImageUrl: "https://documents.iplt20.com/ipl/IPLHeadshot2024/54.png"},
		{Name: "T Natarajan", Role: "BOWL", Nationality: "Indian", Category: "BOWLER", BasePrice: 50, ImageUrl: "https://documents.iplt20.com/ipl/IPLHeadshot2024/54.png"},
		{Name: "Bhuvneshwar Kumar", Role: "BOWL", Nationality: "Indian", Category: "BOWLER", BasePrice: 100, ImageUrl: "https://documents.iplt20.com/ipl/IPLHeadshot2024/54.png"},
		{Name: "Tim David", Role: "BAT", Nationality: "Overseas", Category: "BATSMAN", BasePrice: 100, ImageUrl: "https://documents.iplt20.com/ipl/IPLHeadshot2024/54.png"},
		{Name: "Gerald Coetzee", Role: "BOWL", Nationality: "Overseas", Category: "BOWLER", BasePrice: 150, ImageUrl: "https://documents.iplt20.com/ipl/IPLHeadshot2024/54.png"},
		{Name: "Tilak Varma", Role: "BAT", Nationality: "Indian", Category: "BATSMAN", BasePrice: 50, ImageUrl: "https://documents.iplt20.com/ipl/IPLHeadshot2024/54.png"},
		{Name: "Nehal Wadhera", Role: "BAT", Nationality: "Indian", Category: "BATSMAN", BasePrice: 20, ImageUrl: "https://documents.iplt20.com/ipl/IPLHeadshot2024/54.png"},
		{Name: "Dhruv Jurel", Role: "WK-BAT", Nationality: "Indian", Category: "WK", BasePrice: 50, ImageUrl: "https://documents.iplt20.com/ipl/IPLHeadshot2024/54.png"},
		{Name: "Riyan Parag", Role: "ALL", Nationality: "Indian", Category: "ALLROUNDER", BasePrice: 50, ImageUrl: "https://documents.iplt20.com/ipl/IPLHeadshot2024/54.png"},
		{Name: "Ravichandran Ashwin", Role: "ALL", Nationality: "Indian", Category: "ALLROUNDER", BasePrice: 100, ImageUrl: "https://documents.iplt20.com/ipl/IPLHeadshot2024/54.png"},
		{Name: "Jos Buttler", Role: "WK-BAT", Nationality: "Overseas", Category: "WK", BasePrice: 200, ImageUrl: "https://documents.iplt20.com/ipl/IPLHeadshot2024/54.png"},
		{Name: "Shimron Hetmyer", Role: "BAT", Nationality: "Overseas", Category: "BATSMAN", BasePrice: 150, ImageUrl: "https://documents.iplt20.com/ipl/IPLHeadshot2024/54.png"},
		{Name: "Jason Holder", Role: "ALL", Nationality: "Overseas", Category: "ALLROUNDER", BasePrice: 150, ImageUrl: "https://documents.iplt20.com/ipl/IPLHeadshot2024/54.png"},
		{Name: "Rovman Powell", Role: "BAT", Nationality: "Overseas", Category: "BATSMAN", BasePrice: 50, ImageUrl: "https://documents.iplt20.com/ipl/IPLHeadshot2024/54.png"},
		{Name: "Shreyas Iyer", Role: "BAT", Nationality: "Indian", Category: "BATSMAN", BasePrice: 150, ImageUrl: "https://documents.iplt20.com/ipl/IPLHeadshot2024/54.png"},
		{Name: "Nitish Rana", Role: "ALL", Nationality: "Indian", Category: "ALLROUNDER", BasePrice: 50, ImageUrl: "https://documents.iplt20.com/ipl/IPLHeadshot2024/54.png"},
		{Name: "Ramandeep Singh", Role: "ALL", Nationality: "Indian", Category: "ALLROUNDER", BasePrice: 20, ImageUrl: "https://documents.iplt20.com/ipl/IPLHeadshot2024/54.png"},
		{Name: "Khaleel Ahmed", Role: "BOWL", Nationality: "Indian", Category: "BOWLER", BasePrice: 50, ImageUrl: "https://documents.iplt20.com/ipl/IPLHeadshot2024/54.png"},
		{Name: "Mukesh Kumar", Role: "BOWL", Nationality: "Indian", Category: "BOWLER", BasePrice: 50, ImageUrl: "https://documents.iplt20.com/ipl/IPLHeadshot2024/54.png"},
		{Name: "Harshal Patel", Role: "BOWL", Nationality: "Indian", Category: "BOWLER", BasePrice: 100, ImageUrl: "https://documents.iplt20.com/ipl/IPLHeadshot2024/54.png"},
		{Name: "Jitesh Sharma", Role: "WK-BAT", Nationality: "Indian", Category: "WK", BasePrice: 50, ImageUrl: "https://documents.iplt20.com/ipl/IPLHeadshot2024/54.png"},
		{Name: "Rahul Chahar", Role: "BOWL", Nationality: "Indian", Category: "BOWLER", BasePrice: 50, ImageUrl: "https://documents.iplt20.com/ipl/IPLHeadshot2024/54.png"},
		{Name: "Prabhsimran Singh", Role: "WK-BAT", Nationality: "Indian", Category: "WK", BasePrice: 20, ImageUrl: "https://documents.iplt20.com/ipl/IPLHeadshot2024/54.png"},
		{Name: "Jonny Bairstow", Role: "WK-BAT", Nationality: "Overseas", Category: "WK", BasePrice: 150, ImageUrl: "https://documents.iplt20.com/ipl/IPLHeadshot2024/54.png"},
		{Name: "Will Jacks", Role: "ALL", Nationality: "Overseas", Category: "ALLROUNDER", BasePrice: 100, ImageUrl: "https://documents.iplt20.com/ipl/IPLHeadshot2024/54.png"},

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
