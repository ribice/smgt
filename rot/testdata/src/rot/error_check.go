package rot

import "context"

func serviceValidation() {
	client := &Client{}
	svc, err := NewService(client)
	if err != nil {
		return
	}
	_, err = svc.Validate(context.Background(), "maybe@example.com")
	if err != nil {
		return
	}
}

type Client struct{}
type Service struct{}

func NewService(client *Client) (*Service, error) {
	return &Service{}, nil
}

func (s *Service) Validate(ctx context.Context, email string) (bool, error) {
	return true, nil
}

