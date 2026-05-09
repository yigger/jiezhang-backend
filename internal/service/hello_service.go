package service

import "fmt"

type Greeter interface {
	Hello(name string) string
}

type HelloService struct{}

func NewHelloService() HelloService {
	return HelloService{}
}

func (s HelloService) Hello(name string) string {
	return fmt.Sprintf("hello, %s", name)
}
