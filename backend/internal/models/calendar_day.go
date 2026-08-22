//backend/internal/models/calendar_day.go

package models

type CalendarDay struct {
  ID     uint `gorm:"primaryKey"	json:"id"`
  Year   int  `gorm:"not null"    json:"year"`
  Month  int  `gorm:"not null"    json:"month"`
  Day    int  `gorm:"not null"    json:"day"`
  TaskID uint `gorm:"index"       json:"taskId"`
}
