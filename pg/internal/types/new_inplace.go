package types

import (
	"context"
	"errors"

	"github.com/Deimvis/go-ext/go1.25/xcheck/xmust"
	"github.com/Deimvis/go-ext/go1.25/xoptional"
)

func NewConnReflectInplace(conn Conn, meta MutableConnMeta) ConnReflect {
	return connReflect{Conn: conn, MutableConnMeta: meta}
}

func NewConnMetaInplace(mode ConnMode, isOneTime xoptional.T[bool]) *StructConnMeta {
	return &StructConnMeta{Mode_: xoptional.New(mode), IsOneTime_: isOneTime}
}

func NewOwnedConnInplace(freeFn ConnFreeFn) OwnedConn {
	return ownedConn{freeFn: freeFn}
}

func NewConnOwnershipInplace(meta MutableConnMeta) ConnOwnership {
	return connOwnership{meta: meta}
}

type connReflect struct {
	Conn
	MutableConnMeta
}

type StructConnMeta struct {
	Mode_      xoptional.T[ConnMode]
	IsOneTime_ xoptional.T[bool]
	IsLazy_    xoptional.T[bool]

	OwnershipTaken_ xoptional.T[bool]
	OwnedConn_      OwnedConn
}

// interface guard
var _ ConnMeta = (*StructConnMeta)(nil)
var _ MutableConnMeta = (*StructConnMeta)(nil)

func (cm StructConnMeta) Mode() xoptional.T[ConnMode] {
	return cm.Mode_
}

func (cm StructConnMeta) OwnershipTaken() xoptional.T[bool] {
	return cm.OwnershipTaken_
}

func (cm StructConnMeta) IsOneTime() xoptional.T[bool] {
	return cm.IsOneTime_
}

func (cm StructConnMeta) IsLazy() xoptional.T[bool] {
	return cm.IsLazy_
}

func (cm *StructConnMeta) TakeOwnership() (OwnedConn, error) {
	if !cm.OwnershipTaken_.HasValue() {
		return nil, errors.New("ownership status is unknown")
	} else if cm.OwnershipTaken_.Value() {
		return nil, errors.New("ownership is already taken")
	}
	cm.OwnershipTaken_.SetValue(true)
	return cm.OwnedConn_, nil
}

type ownedConn struct {
	freeFn ConnFreeFn
}

func (oc ownedConn) FreeConn(ctx context.Context) error {
	return oc.freeFn(ctx)
}

type connOwnership struct {
	meta MutableConnMeta
}

func (co connOwnership) Take() (OwnedConn, error) {
	return co.meta.TakeOwnership()
}

func (co connOwnership) MustTake() OwnedConn {
	return xmust.Do(co.meta.TakeOwnership())
}
