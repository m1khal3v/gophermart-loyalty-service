package client

import "fmt"

type ErrUnexpectedStatus struct {
	Status int
}

func (err ErrUnexpectedStatus) Error() string {
	return fmt.Sprintf("unexpected status code: %d", err.Status)
}

func newErrUnexpectedStatus(status int) ErrUnexpectedStatus {
	return ErrUnexpectedStatus{
		Status: status,
	}
}

type ErrInvalidAddress struct {
	Address string
}

func (err ErrInvalidAddress) Error() string {
	return fmt.Sprintf("invalid address: %s", err.Address)
}

func newErrInvalidAddress(address string) ErrInvalidAddress {
	return ErrInvalidAddress{
		Address: address,
	}
}
