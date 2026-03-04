package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"tuno_backend/internal/domain"

	"github.com/redis/go-redis/v9"
)

type CachedGroupRepository struct {
	primaryRepo domain.GroupRepository
	redisClient *redis.Client
}

func NewCachedGroupRepository(primaryRepo domain.GroupRepository, redisClient *redis.Client) *CachedGroupRepository {
	return &CachedGroupRepository{
		primaryRepo: primaryRepo,
		redisClient: redisClient,
	}
}

func (r *CachedGroupRepository) GetGroupHome(groupID, userID string) (*domain.GroupHomeResponse, error) {
	ctx := context.Background()
	cacheKey := fmt.Sprintf("group_home:%s:%s", groupID, userID)

	// Try to get from cache first
	cachedData, err := r.redisClient.Get(ctx, cacheKey).Bytes()
	if err == nil {
		var response domain.GroupHomeResponse
		if err := json.Unmarshal(cachedData, &response); err == nil {
			return &response, nil
		}
	}

	// Cache miss or invalid data - get from primary repository
	response, err := r.primaryRepo.GetGroupHome(groupID, userID)
	if err != nil {
		return nil, err
	}

	// Cache the response for 5 minutes
	if data, err := json.Marshal(response); err == nil {
		r.redisClient.Set(ctx, cacheKey, data, 5*time.Minute)
	}

	return response, nil
}

func (r *CachedGroupRepository) InvalidateGroupHomeCache(groupID, userID string) error {
	ctx := context.Background()
	cacheKey := fmt.Sprintf("group_home:%s:%s", groupID, userID)
	return r.redisClient.Del(ctx, cacheKey).Err()
}

func (r *CachedGroupRepository) InvalidateAllGroupHomeCache(groupID string) error {
	ctx := context.Background()
	pattern := fmt.Sprintf("group_home:%s:*", groupID)

	iter := r.redisClient.Scan(ctx, 0, pattern, 0).Iterator()
	for iter.Next(ctx) {
		r.redisClient.Del(ctx, iter.Val())
	}
	return iter.Err()
}

// Implement all other GroupRepository methods by delegating to primaryRepo
func (r *CachedGroupRepository) Create(group *domain.Group) error {
	return r.primaryRepo.Create(group)
}

func (r *CachedGroupRepository) AddMember(member *domain.GroupMember) error {
	return r.primaryRepo.AddMember(member)
}

func (r *CachedGroupRepository) GetMember(groupID, userID string) (*domain.GroupMember, error) {
	return r.primaryRepo.GetMember(groupID, userID)
}

func (r *CachedGroupRepository) FindByID(id string) (*domain.Group, error) {
	return r.primaryRepo.FindByID(id)
}

func (r *CachedGroupRepository) CreateGroupWithMembers(group *domain.Group, members []domain.GroupMember, initialMessage *domain.Message) error {
	return r.primaryRepo.CreateGroupWithMembers(group, members, initialMessage)
}

func (r *CachedGroupRepository) AddMemberWithSystemMessage(member *domain.GroupMember, systemMessage *domain.Message) error {
	return r.primaryRepo.AddMemberWithSystemMessage(member, systemMessage)
}

func (r *CachedGroupRepository) GetGroupsByUserID(userID string) ([]domain.GroupWithRole, error) {
	return r.primaryRepo.GetGroupsByUserID(userID)
}

func (r *CachedGroupRepository) GetGroupDetail(groupID, userID string) (*domain.GroupDetail, error) {
	return r.primaryRepo.GetGroupDetail(groupID, userID)
}

func (r *CachedGroupRepository) GetGroupMemberIDs(groupID string) ([]string, error) {
	return r.primaryRepo.GetGroupMemberIDs(groupID)
}

func (r *CachedGroupRepository) GetGroupMembers(groupID string) ([]domain.GroupMemberView, error) {
	return r.primaryRepo.GetGroupMembers(groupID)
}

func (r *CachedGroupRepository) GetGroupMembersPaginated(groupID string, limit, offset int) ([]domain.GroupMemberView, int, error) {
	return r.primaryRepo.GetGroupMembersPaginated(groupID, limit, offset)
}

func (r *CachedGroupRepository) HasCommonActiveGroup(user1ID, user2ID string) (bool, error) {
	return r.primaryRepo.HasCommonActiveGroup(user1ID, user2ID)
}

func (r *CachedGroupRepository) GetGroupSchedule(groupID string, filterStatus *domain.RoundStatus) (*domain.GroupScheduleResponse, error) {
	return r.primaryRepo.GetGroupSchedule(groupID, filterStatus)
}
