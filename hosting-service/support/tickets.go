package support

import (
	"fmt"
	"time"
)

// TicketPriority represents the priority level of a support ticket
type TicketPriority string

const (
	PriorityLow      TicketPriority = "low"
	PriorityMedium   TicketPriority = "medium"
	PriorityHigh     TicketPriority = "high"
	PriorityCritical TicketPriority = "critical"
)

// TicketStatus represents the current status of a support ticket
type TicketStatus string

const (
	StatusOpen       TicketStatus = "open"
	StatusInProgress TicketStatus = "in_progress"
	StatusWaiting    TicketStatus = "waiting_customer"
	StatusResolved   TicketStatus = "resolved"
	StatusClosed     TicketStatus = "closed"
)

// Ticket represents a customer support ticket
type Ticket struct {
	ID          uint           `json:"id"`
	CustomerID  uint           `json:"customer_id"`
	Title       string         `json:"title"`
	Description string         `json:"description"`
	Category    string         `json:"category"`
	Priority    TicketPriority `json:"priority"`
	Status      TicketStatus   `json:"status"`
	AssignedTo  *uint          `json:"assigned_to,omitempty"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	ResolvedAt  *time.Time     `json:"resolved_at,omitempty"`

	// Relationships
	Messages []TicketMessage `json:"messages,omitempty"`
}

// TicketMessage represents a message within a support ticket
type TicketMessage struct {
	ID         uint      `json:"id"`
	TicketID   uint      `json:"ticket_id"`
	AuthorID   uint      `json:"author_id"`
	AuthorType string    `json:"author_type"` // "customer" or "support"
	Message    string    `json:"message"`
	IsInternal bool      `json:"is_internal"`
	CreatedAt  time.Time `json:"created_at"`
}

// SupportService handles support ticket operations
type SupportService struct {
	tickets map[uint]*Ticket
	nextID  uint
}

func NewSupportService() *SupportService {
	return &SupportService{
		tickets: make(map[uint]*Ticket),
		nextID:  1,
	}
}

// CreateTicket creates a new support ticket
func (s *SupportService) CreateTicket(customerID uint, title, description, category string, priority TicketPriority) (*Ticket, error) {
	ticket := &Ticket{
		ID:          s.nextID,
		CustomerID:  customerID,
		Title:       title,
		Description: description,
		Category:    category,
		Priority:    priority,
		Status:      StatusOpen,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Messages:    []TicketMessage{},
	}

	s.tickets[s.nextID] = ticket
	s.nextID++

	return ticket, nil
}

// GetTicket retrieves a ticket by ID
func (s *SupportService) GetTicket(ticketID uint) (*Ticket, error) {
	ticket, exists := s.tickets[ticketID]
	if !exists {
		return nil, fmt.Errorf("ticket not found")
	}
	return ticket, nil
}

// GetCustomerTickets retrieves all tickets for a customer
func (s *SupportService) GetCustomerTickets(customerID uint) ([]*Ticket, error) {
	var customerTickets []*Ticket
	for _, ticket := range s.tickets {
		if ticket.CustomerID == customerID {
			customerTickets = append(customerTickets, ticket)
		}
	}
	return customerTickets, nil
}

// AddMessage adds a message to a ticket
func (s *SupportService) AddMessage(ticketID, authorID uint, authorType, message string, isInternal bool) (*TicketMessage, error) {
	ticket, exists := s.tickets[ticketID]
	if !exists {
		return nil, fmt.Errorf("ticket not found")
	}

	msg := TicketMessage{
		ID:         uint(len(ticket.Messages) + 1),
		TicketID:   ticketID,
		AuthorID:   authorID,
		AuthorType: authorType,
		Message:    message,
		IsInternal: isInternal,
		CreatedAt:  time.Now(),
	}

	ticket.Messages = append(ticket.Messages, msg)
	ticket.UpdatedAt = time.Now()

	// Update status if customer responds
	if authorType == "customer" && ticket.Status == StatusWaiting {
		ticket.Status = StatusInProgress
	}

	return &msg, nil
}

// UpdateTicketStatus updates the status of a ticket
func (s *SupportService) UpdateTicketStatus(ticketID uint, status TicketStatus) error {
	ticket, exists := s.tickets[ticketID]
	if !exists {
		return fmt.Errorf("ticket not found")
	}

	ticket.Status = status
	ticket.UpdatedAt = time.Now()

	if status == StatusResolved || status == StatusClosed {
		now := time.Now()
		ticket.ResolvedAt = &now
	}

	return nil
}

// AssignTicket assigns a ticket to a support agent
func (s *SupportService) AssignTicket(ticketID, agentID uint) error {
	ticket, exists := s.tickets[ticketID]
	if !exists {
		return fmt.Errorf("ticket not found")
	}

	ticket.AssignedTo = &agentID
	ticket.UpdatedAt = time.Now()

	if ticket.Status == StatusOpen {
		ticket.Status = StatusInProgress
	}

	return nil
}

// GetTicketsByStatus retrieves tickets filtered by status
func (s *SupportService) GetTicketsByStatus(status TicketStatus) ([]*Ticket, error) {
	var filteredTickets []*Ticket
	for _, ticket := range s.tickets {
		if ticket.Status == status {
			filteredTickets = append(filteredTickets, ticket)
		}
	}
	return filteredTickets, nil
}

// GetTicketsByPriority retrieves tickets filtered by priority
func (s *SupportService) GetTicketsByPriority(priority TicketPriority) ([]*Ticket, error) {
	var filteredTickets []*Ticket
	for _, ticket := range s.tickets {
		if ticket.Priority == priority {
			filteredTickets = append(filteredTickets, ticket)
		}
	}
	return filteredTickets, nil
}

// GetTicketStats returns basic statistics about tickets
func (s *SupportService) GetTicketStats() map[string]int {
	stats := map[string]int{
		"total":       len(s.tickets),
		"open":        0,
		"in_progress": 0,
		"resolved":    0,
		"closed":      0,
		"critical":    0,
	}

	for _, ticket := range s.tickets {
		switch ticket.Status {
		case StatusOpen:
			stats["open"]++
		case StatusInProgress:
			stats["in_progress"]++
		case StatusResolved:
			stats["resolved"]++
		case StatusClosed:
			stats["closed"]++
		}

		if ticket.Priority == PriorityCritical {
			stats["critical"]++
		}
	}

	return stats
}
