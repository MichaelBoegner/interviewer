package user

import "log"

func GetUsers(repo *Repository) (*Users, error) {
	userMap := make(map[int]User)
	users := &Users{
		Users: userMap,
	}
	usersReturned, err := repo.GetUsers(users)
	if err != nil {
		log.Printf("GetUsers from database failed due to: %v", err)
		return nil, err
	}
	return usersReturned, nil
}
