package domain

type Location struct {
	Lat float64 `form:"lat" json:"lat" example:"59.934280" swagger:"description='Широта'"`
	Lon float64 `form:"lon" json:"lon" example:"30.335098" swagger:"description='Долгота'"`
}
