package auth

//go:generate cmd /c "if exist mocks rmdir /s /q mocks && mkdir mocks"
//go:generate minimock -i AuthService -o ./mocks/ -s "_minimock.go"
//go:generate minimock -i UsersReader -o ./mocks/ -s "_minimock.go"
