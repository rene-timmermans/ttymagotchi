package app

import "time"

type Tamagotchi struct {
	Species          string
	HealthyTicks     int
	Cleanliness      int
	GroomTicks       int
	DiaperDayTicks   int
	DiaperEventCount int
	DiaperNextEvent  int
	DiaperEvents     [5]int
	Diaper           int
	LastChange       time.Time
	Grooming         int
	CareTicks        int
	LastWash         time.Time
	LastBrush        time.Time
	Name             string
	Age              int
	Hunger           int
	Happiness        int
	Health           int
	Energy           int
	Weight           float64
	Stage            string
	Created          time.Time
	LastFed          time.Time
	LastPlay         time.Time
	LastSleep        time.Time
	IsAlive          bool
}

type GameEvent struct {
	MessageKey  string
	MessageArgs []any
	Type        string
	Message     string
	Timestamp   time.Time
}

type Food struct {
	Name       string
	Nutrition  int
	Happiness  int
	Energy     int
	WeightGain float64
}
