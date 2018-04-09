// Copyright 2018 The Loopix-Messaging Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package logging

import (
	"github.com/sirupsen/logrus"
	"runtime"
)

var baseLogger = logrus.New()

type AnonymousMessagingLogger struct {
	*logrus.Entry
}

func (l *AnonymousMessagingLogger) WithField(key string, value interface{}) *AnonymousMessagingLogger {
	return &AnonymousMessagingLogger{l.Entry.WithField(key, value)}
}

func (l *AnonymousMessagingLogger) WithFields(fields logrus.Fields) *AnonymousMessagingLogger {
	return &AnonymousMessagingLogger{l.Entry.WithFields(fields)}
}

func PackageLogger() *AnonymousMessagingLogger {
	_, filename, _, _ := runtime.Caller(1)
	return &AnonymousMessagingLogger{baseLogger.WithField("prefix", filename)}
}

func PackageLoggerWithField(key, value string) *AnonymousMessagingLogger {
	return &AnonymousMessagingLogger{baseLogger.WithField(key, value)}
}
