package domain

import "time"

type User struct {
	ID          *string    `form:"id" json:"id,omitempty" example:"507f1f77bcf86cd799439011" swagger:"description='Уникальный идентификатор пользователя'"`
	Login       *string    `form:"login" json:"login,omitempty" example:"john_doe" swagger:"description='Логин пользователя'"`
	Username    *string    `form:"username" json:"username,omitempty" example:"John Doe" swagger:"description='Имя пользователя'"`
	Password    *string    `form:"password" json:"password,omitempty" example:"secret123" swagger:"description='Пароль пользователя '"`
	Description *string    `form:"description" json:"description,omitempty" example:"Программист из Санкт-Петербурга" swagger:"description='Описание пользователя'"`
	Comment     *string    `form:"comment" json:"comment,omitempty" example:"Важный клиент" swagger:"description='Комментарии о пользователе (заметка админа)'"`
	RegDate     *time.Time `form:"reg_date" json:"reg_date,omitempty" example:"2023-01-15T12:34:56Z" swagger:"description='Дата регистрации'"`
	Location    *Location  `form:"location" json:"location,omitempty" swagger:"description='Геолокация пользователя'"`
	SocialNet   *string    `form:"social_net" json:"social_net,omitempty" example:"MAX" swagger:"description='Название соц. сети, строго определенное'"`
}

type UserFilter struct {
	// Полнотекстовый поиск
	Search *string `form:"q" json:"search,omitempty" example:"john_doe" swagger:"description='Поисковый запрос (имя, логин, описание, комментарий)'"`

	// Фильтры по дате
	DateFrom *time.Time `form:"date_from" json:"date_from,omitempty" example:"2024-01-01T00:00:00Z" swagger:"description='Дата регистрации от (RFC3339)', format='date-time'"`
	DateTo   *time.Time `form:"date_to" json:"date_to,omitempty" example:"2024-12-31T23:59:59Z" swagger:"description='Дата регистрации до (RFC3339)', format='date-time'"`

	// Гео-фильтр
	Lat      *float64 `form:"lat" json:"lat,omitempty" example:"52.52" swagger:"description='Широта для поиска'"`
	Lon      *float64 `form:"lon" json:"lon,omitempty" example:"13.41" swagger:"description='Долгота для поиска'"`
	Distance *string  `form:"radius" json:"radius,omitempty" example:"1000" swagger:"description='Максимально расстояние между пользователем и заданной точки', default='1km', enum='100m,500m,1km,5km,10km'"`

	// Социальные сети
	SocialType *string `form:"social_net" json:"social_net,omitempty" example:"facebook" swagger:"description='Тип соц. сети'"`

	// Сортировка
	SortBy    *string `form:"sort_by" json:"sort_by,omitempty" example:"login" swagger:"description='Поле сортировки (login, reg_date)', enum='login,reg_date'"`
	SortOrder *string `form:"sort_order" json:"sort_order,omitempty" example:"desc" swagger:"description='Порядок сортировки (asc, desc)', enum='asc,desc'"`

	// Пагинация
	Page *int `form:"page" json:"page,omitempty" example:"1" swagger:"description='Номер страницы', default='1'"`
	Size *int `form:"size" json:"size,omitempty" example:"10" swagger:"description='Лимит записей', default='10'"`
}
