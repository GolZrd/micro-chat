package auth

//go:generate cmd /c "if exist mocks rmdir /s /q mocks && mkdir mocks"
//go:generate minimock -i AuthRepository -o ./mocks/ -s "_minimock.go"
