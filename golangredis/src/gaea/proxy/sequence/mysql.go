// Copyright 2019 The Gaea Authors. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package sequence

import (
	"sync"

	"github.com/XiaoMi/Gaea/backend"
)

// MySQLSequence struct of sequence number with specific sequence name
type MySQLSequence struct {
	slice   *backend.Slice
	pkName  string
	seqName string
	lock    *sync.Mutex
	curr    int64
	max     int64
	sql     string
}

// NewMySQLSequence init sequence item
// TODO: 直接注入slice需要考虑关闭的问题, 目前是在Namespace中管理Slice的关闭的. 如果单独使用MySQLSequence, 需要注意.
func NewMySQLSequence(slice *backend.Slice, seqName, pkName string) *MySQLSequence {
	t := &MySQLSequence{
		slice:   slice,
		seqName: seqName,
		pkName:  pkName,
		lock:    new(sync.Mutex),
		curr:    0,
		max:     0,
		sql:     "SELECT mycat_seq_nextval('" + seqName + "') as seq_val",
	}
	return t
}

// NextSeq get next sequence number
func (s *MySQLSequence) NextSeq() (int64, error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.curr++
	if s.curr > s.max {
		err := s.getSeqFromDB()
		if err != nil {
			return 0, err
		}
	}
	t := s.curr
	s.curr++
	return t, nil
}

// GetPKName return sequence column
func (s *MySQLSequence) GetPKName() string {
	return s.pkName
}

func (s *MySQLSequence) getSeqFromDB() error {
	conn, err := s.slice.GetMasterConn()
	if err != nil {
		return err
	}
	defer conn.Recycle()

	err = conn.UseDB("mycat")
	if err != nil {
		return err
	}

	r, err := conn.Execute(s.sql)
	if err != nil {
		return err
	}
	curr, err := r.Resultset.GetInt(0, 0)
	if err != nil {
		return nil
	}

	incr, err := r.Resultset.GetInt(0, 1)
	if err != nil {
		return nil
	}
	s.max = curr + incr

	return nil
}
