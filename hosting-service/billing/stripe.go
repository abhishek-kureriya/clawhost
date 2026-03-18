package billing

import (
	"fmt"
	"log"

	"github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/customer"
	"github.com/stripe/stripe-go/v76/paymentintent"
	"github.com/stripe/stripe-go/v76/price"
	"github.com/stripe/stripe-go/v76/product"
	"github.com/stripe/stripe-go/v76/subscription"
	"github.com/stripe/stripe-go/v76/webhook"
)

// BillingService handles all Stripe-related operations for the hosting service
type BillingService struct {
	secretKey string
}

type CustomerData struct {
	Email       string
	Name        string
	CompanyName string
}

type SubscriptionData struct {
	CustomerID string
	PriceID    string
	PlanName   string
}

func NewBillingService(secretKey string) *BillingService {
	stripe.Key = secretKey
	return &BillingService{
		secretKey: secretKey,
	}
}

func (s *BillingService) CreateCustomer(customerData CustomerData) (*stripe.Customer, error) {
	params := &stripe.CustomerParams{
		Email: stripe.String(customerData.Email),
		Name:  stripe.String(customerData.Name),
	}

	if customerData.CompanyName != "" {
		params.Description = stripe.String(fmt.Sprintf("Company: %s", customerData.CompanyName))
	}

	cust, err := customer.New(params)
	if err != nil {
		return nil, fmt.Errorf("failed to create Stripe customer: %w", err)
	}

	log.Printf("Created Stripe customer: %s (%s)", cust.ID, customerData.Email)
	return cust, nil
}

func (s *BillingService) CreateSubscription(subscriptionData SubscriptionData) (*stripe.Subscription, error) {
	params := &stripe.SubscriptionParams{
		Customer: stripe.String(subscriptionData.CustomerID),
		Items: []*stripe.SubscriptionItemsParams{
			{
				Price: stripe.String(subscriptionData.PriceID),
			},
		},
		PaymentBehavior: stripe.String("default_incomplete"),
		PaymentSettings: &stripe.SubscriptionPaymentSettingsParams{
			SaveDefaultPaymentMethod: stripe.String("on_subscription"),
		},
	}

	params.AddExpand("latest_invoice.payment_intent")

	sub, err := subscription.New(params)
	if err != nil {
		return nil, fmt.Errorf("failed to create subscription: %w", err)
	}

	log.Printf("Created subscription: %s for customer: %s", sub.ID, subscriptionData.CustomerID)
	return sub, nil
}

func (s *BillingService) CreatePaymentIntent(amount int64, currency string, customerID string) (*stripe.PaymentIntent, error) {
	params := &stripe.PaymentIntentParams{
		Amount:   stripe.Int64(amount),
		Currency: stripe.String(currency),
		Customer: stripe.String(customerID),
		AutomaticPaymentMethods: &stripe.PaymentIntentAutomaticPaymentMethodsParams{
			Enabled: stripe.Bool(true),
		},
	}

	pi, err := paymentintent.New(params)
	if err != nil {
		return nil, fmt.Errorf("failed to create payment intent: %w", err)
	}

	return pi, nil
}

func (s *BillingService) GetSubscription(subscriptionID string) (*stripe.Subscription, error) {
	sub, err := subscription.Get(subscriptionID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get subscription: %w", err)
	}

	return sub, nil
}

func (s *BillingService) CancelSubscription(subscriptionID string) (*stripe.Subscription, error) {
	params := &stripe.SubscriptionParams{
		CancelAtPeriodEnd: stripe.Bool(true),
	}

	sub, err := subscription.Update(subscriptionID, params)
	if err != nil {
		return nil, fmt.Errorf("failed to cancel subscription: %w", err)
	}

	log.Printf("Cancelled subscription: %s", subscriptionID)
	return sub, nil
}

func (s *BillingService) ConstructWebhookEvent(body []byte, signature string, endpointSecret string) (stripe.Event, error) {
	event, err := webhook.ConstructEvent(body, signature, endpointSecret)
	if err != nil {
		return event, fmt.Errorf("failed to construct webhook event: %w", err)
	}

	return event, nil
}

// Helper function to get price IDs for different plans
func (s *BillingService) GetPriceIDForPlan(planName string) string {
	priceIDs := map[string]string{
		"starter":      "price_starter_monthly",      // Replace with actual Stripe price ID
		"professional": "price_professional_monthly", // Replace with actual Stripe price ID
		"enterprise":   "price_enterprise_monthly",   // Replace with actual Stripe price ID
	}

	return priceIDs[planName]
}

// SetupProducts creates Stripe products for ClawHost plans
func (s *BillingService) SetupProducts() error {
	plans := []struct {
		Name        string
		Description string
		Amount      int64 // in cents (EUR)
		PlanType    string
	}{
		{
			Name:        "ClawHost Starter",
			Description: "Perfect for small businesses getting started with AI",
			Amount:      4900, // €49.00
			PlanType:    "starter",
		},
		{
			Name:        "ClawHost Professional",
			Description: "Ideal for growing businesses with multiple AI use cases",
			Amount:      9900, // €99.00
			PlanType:    "professional",
		},
		{
			Name:        "ClawHost Enterprise",
			Description: "For large organizations with complex AI requirements",
			Amount:      19900, // €199.00
			PlanType:    "enterprise",
		},
	}

	for _, plan := range plans {
		// Create product
		productParams := &stripe.ProductParams{
			Name:        stripe.String(plan.Name),
			Description: stripe.String(plan.Description),
		}
		prod, err := product.New(productParams)
		if err != nil {
			log.Printf("Error creating product %s: %v", plan.Name, err)
			continue
		}

		// Create monthly price
		priceParams := &stripe.PriceParams{
			Product:    stripe.String(prod.ID),
			UnitAmount: stripe.Int64(plan.Amount),
			Currency:   stripe.String("eur"),
			Recurring: &stripe.PriceRecurringParams{
				Interval: stripe.String("month"),
			},
		}
		price, err := price.New(priceParams)
		if err != nil {
			log.Printf("Error creating price for %s: %v", plan.Name, err)
			continue
		}

		log.Printf("Created Stripe product: %s with price: %s (%s)", prod.ID, price.ID, plan.PlanType)
	}

	return nil
}
