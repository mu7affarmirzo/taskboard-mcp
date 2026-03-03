package dto

type ListMembersOutput struct {
	Members []MemberItem
}

type MemberItem struct {
	ID       string
	Username string
	FullName string
}
