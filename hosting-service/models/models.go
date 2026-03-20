package models

import "time"

type Customer struct {
	ID               uint      `gorm:"primaryKey" json:"id"`
	Email            string    `gorm:"uniqueIndex;not null" json:"email"`
	Name             string    `json:"name"`
	CompanyName      string    `json:"company_name"`
	StripeCustomerID string    `gorm:"index" json:"stripe_customer_id"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`

	Instances     []Instance     `json:"instances,omitempty"`
	Tickets       []SupportTicket `json:"tickets,omitempty"`
	Subscriptions []Subscription  `json:"subscriptions,omitempty"`
}

type Instance struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	CustomerID      uint      `gorm:"index;not null" json:"customer_id"`
	Name            string    `gorm:"not null" json:"name"`
	Status          string    `gorm:"index" json:"status"`
	Provider        string    `json:"provider"`
	ServerType      string    `json:"server_type"`
	Location        string    `json:"location"`
	PublicIP        string    `json:"public_ip"`
	OpenClawURL     string    `json:"openclaw_url"`
	ExternalJobID   string    `gorm:"index" json:"external_job_id"`
	Configuration   string    `gorm:"type:text" json:"configuration"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type SupportTicket struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	CustomerID  uint      `gorm:"index;not null" json:"customer_id"`
	Title       string    `gorm:"not null" json:"title"`
	Description string    `gorm:"type:text" json:"description"`
	Category    string    `json:"category"`
	Priority    string    `gorm:"index" json:"priority"`
	Status      string    `gorm:"index" json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	Messages []SupportMessage `json:"messages,omitempty"`
}

type SupportMessage struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	TicketID   uint      `gorm:"index;not null" json:"ticket_id"`
	AuthorID   uint      `json:"author_id"`
	AuthorType string    `json:"author_type"`
	Message    string    `gorm:"type:text" json:"message"`
	IsInternal bool      `json:"is_internal"`
	CreatedAt  time.Time `json:"created_at"`
}

type Subscription struct {
	ID                   uint      `gorm:"primaryKey" json:"id"`
	CustomerID           uint      `gorm:"index;not null" json:"customer_id"`
	PlanName             string    `json:"plan_name"`
	Status               string    `gorm:"index" json:"status"`
	StripeSubscriptionID string    `gorm:"uniqueIndex;not null" json:"stripe_subscription_id"`
	StripePriceID        string    `json:"stripe_price_id"`
	CurrentPeriodEnd     time.Time `json:"current_period_end"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}
