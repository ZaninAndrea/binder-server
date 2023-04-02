package achievements

import "github.com/ZaninAndrea/binder-server/internal/mongo"

type Achievement interface {
	ID() string
	Level(user *mongo.User) int
	CurrentLevel(user *mongo.User) int
}

var achievements = []Achievement{
	&TotalRepetitions,
	&ActiveDays,
	&SingleDayRepetitions,
}

type AchievementUpdate struct {
	ID    string
	Level int
}

func UpdateAchievements(user *mongo.User) []AchievementUpdate {
	updates := make([]AchievementUpdate, 0)

	for _, achievement := range achievements {
		currentLevel := achievement.CurrentLevel(user)
		newLevel := achievement.Level(user)

		if newLevel > currentLevel {
			updates = append(updates, AchievementUpdate{
				ID:    achievement.ID(),
				Level: newLevel,
			})
		}
	}

	return updates
}

type StatAchievement struct {
	id                string
	levelRequirements []int
	getValue          func(user *mongo.User) int
	getLevel          func(user *mongo.User) int
}

var TotalRepetitions = StatAchievement{
	id: "totalRepetitions",
	levelRequirements: []int{
		10, 25, 50, 100, 250, 500, 750, 1000, 1500, 2000, 3000, 4000, 5000, 6000, 7000, 8000, 9000, 10000,
	},
	getValue: func(user *mongo.User) int {
		repetitions := 0
		for _, value := range user.Statistics.DailyRepetitions {
			repetitions += value
		}
		return repetitions
	},
	getLevel: func(user *mongo.User) int {
		return user.Achievements.TotalRepetitions
	},
}

var ActiveDays = StatAchievement{
	id: "activeDays",
	levelRequirements: []int{
		3, 7, 14, 30, 50, 100, 150, 200, 300, 365, 500, 730, 1000,
	},
	getValue: func(user *mongo.User) int {
		return len(user.Statistics.DailyRepetitions)
	},
	getLevel: func(user *mongo.User) int {
		return user.Achievements.ActiveDays
	},
}

var SingleDayRepetitions = StatAchievement{
	id: "singleDayRepetitions",
	levelRequirements: []int{
		2, 5, 10, 20, 30, 50, 75, 100, 125, 150, 175, 200,
	},
	getValue: func(user *mongo.User) int {
		maxRepetitions := 0
		for _, value := range user.Statistics.DailyRepetitions {
			if value > maxRepetitions {
				maxRepetitions = value
			}
		}
		return maxRepetitions
	},
	getLevel: func(user *mongo.User) int {
		return user.Achievements.SingleDayRepetitions
	},
}

func (a *StatAchievement) ID() string {
	return a.id
}

func (a *StatAchievement) Level(user *mongo.User) int {
	value := a.getValue(user)

	for i, requirement := range a.levelRequirements {
		if value < requirement {
			return i
		}
	}

	return len(a.levelRequirements)
}

func (a *StatAchievement) CurrentLevel(user *mongo.User) int {
	return a.getLevel(user)
}
