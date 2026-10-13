package mocks

import (
	context "context"

	domain "github.com/janghanul090801/go-backend-clean-architecture-fiber/domain"
	mock "github.com/stretchr/testify/mock"

	uuid "github.com/google/uuid"
)

type LegacyRepository struct {
	mock.Mock
}

func (_m *LegacyRepository) Create(c context.Context, legacy *domain.Legacy, userID *uuid.UUID) (*domain.Legacy, error) {
	ret := _m.Called(c, legacy, userID)

	var r0 *domain.Legacy
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *domain.Legacy, *uuid.UUID) (*domain.Legacy, error)); ok {
		return rf(c, legacy, userID)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *domain.Legacy, *uuid.UUID) *domain.Legacy); ok {
		r0 = rf(c, legacy, userID)
	} else if ret.Get(0) != nil {
		r0 = ret.Get(0).(*domain.Legacy)
	}

	if rf, ok := ret.Get(1).(func(context.Context, *domain.Legacy, *uuid.UUID) error); ok {
		r1 = rf(c, legacy, userID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

func (_m *LegacyRepository) FindAllByUserID(c context.Context, userID *uuid.UUID, offset int, limit int) ([]*domain.Legacy, int, error) {
	ret := _m.Called(c, userID, offset, limit)

	var r0 []*domain.Legacy
	var r1 int
	var r2 error
	if rf, ok := ret.Get(0).(func(context.Context, *uuid.UUID, int, int) ([]*domain.Legacy, int, error)); ok {
		return rf(c, userID, offset, limit)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *uuid.UUID, int, int) []*domain.Legacy); ok {
		r0 = rf(c, userID, offset, limit)
	} else if ret.Get(0) != nil {
		r0 = ret.Get(0).([]*domain.Legacy)
	}

	if rf, ok := ret.Get(1).(func(context.Context, *uuid.UUID, int, int) int); ok {
		r1 = rf(c, userID, offset, limit)
	} else {
		r1 = ret.Get(1).(int)
	}

	if rf, ok := ret.Get(2).(func(context.Context, *uuid.UUID, int, int) error); ok {
		r2 = rf(c, userID, offset, limit)
	} else {
		r2 = ret.Error(2)
	}

	return r0, r1, r2
}

func (_m *LegacyRepository) FindByID(c context.Context, id *uuid.UUID, userID *uuid.UUID) (*domain.Legacy, error) {
	ret := _m.Called(c, id, userID)

	var r0 *domain.Legacy
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *uuid.UUID, *uuid.UUID) (*domain.Legacy, error)); ok {
		return rf(c, id, userID)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *uuid.UUID, *uuid.UUID) *domain.Legacy); ok {
		r0 = rf(c, id, userID)
	} else if ret.Get(0) != nil {
		r0 = ret.Get(0).(*domain.Legacy)
	}

	if rf, ok := ret.Get(1).(func(context.Context, *uuid.UUID, *uuid.UUID) error); ok {
		r1 = rf(c, id, userID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

func (_m *LegacyRepository) Update(c context.Context, id *uuid.UUID, name string, userID *uuid.UUID) (*domain.Legacy, error) {
	ret := _m.Called(c, id, name, userID)

	var r0 *domain.Legacy
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *uuid.UUID, string, *uuid.UUID) (*domain.Legacy, error)); ok {
		return rf(c, id, name, userID)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *uuid.UUID, string, *uuid.UUID) *domain.Legacy); ok {
		r0 = rf(c, id, name, userID)
	} else if ret.Get(0) != nil {
		r0 = ret.Get(0).(*domain.Legacy)
	}

	if rf, ok := ret.Get(1).(func(context.Context, *uuid.UUID, string, *uuid.UUID) error); ok {
		r1 = rf(c, id, name, userID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

func (_m *LegacyRepository) Delete(c context.Context, id *uuid.UUID, userID *uuid.UUID) error {
	ret := _m.Called(c, id, userID)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *uuid.UUID, *uuid.UUID) error); ok {
		r0 = rf(c, id, userID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

func (_m *LegacyRepository) CreateAccount(c context.Context, account *domain.Account, legacyID *uuid.UUID, userID *uuid.UUID) (*domain.Account, error) {
	ret := _m.Called(c, account, legacyID, userID)

	var r0 *domain.Account
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *domain.Account, *uuid.UUID, *uuid.UUID) (*domain.Account, error)); ok {
		return rf(c, account, legacyID, userID)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *domain.Account, *uuid.UUID, *uuid.UUID) *domain.Account); ok {
		r0 = rf(c, account, legacyID, userID)
	} else if ret.Get(0) != nil {
		r0 = ret.Get(0).(*domain.Account)
	}

	if rf, ok := ret.Get(1).(func(context.Context, *domain.Account, *uuid.UUID, *uuid.UUID) error); ok {
		r1 = rf(c, account, legacyID, userID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

func (_m *LegacyRepository) DeleteAccount(c context.Context, accountID *uuid.UUID, legacyID *uuid.UUID, userID *uuid.UUID) error {
	ret := _m.Called(c, accountID, legacyID, userID)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *uuid.UUID, *uuid.UUID, *uuid.UUID) error); ok {
		r0 = rf(c, accountID, legacyID, userID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

func NewLegacyRepository(t interface {
	mock.TestingT
	Cleanup(func())
}) *LegacyRepository {
	mock := &LegacyRepository{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
