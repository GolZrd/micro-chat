package user

//go:generate cmd /c "if exist mocks rmdir /s /q mocks && mkdir mocks"
//go:generate minimock -i UserRepository -o ./mocks/ -s "_minimock.go"
