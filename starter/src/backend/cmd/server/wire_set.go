package main

import (
	"github.com/google/wire"
	"github.com/kibu-sh/kibu/pkg/wireset"
	_ "github.com/lib/pq"
	"kibu.sh/starter/src/backend/kibugen"
)

// ignore unused
// nolint:deadcode,unused
var wireSet = wire.NewSet(
	wireset.DefaultSet,
	kibugen.WireSet,
)
