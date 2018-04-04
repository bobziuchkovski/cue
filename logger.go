// Copyright (c) 2016 Bob Ziuchkovski
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package cue

import (
	"errors"
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

var (
	// We use the internal context and logger to report our own internal
	// events, such as collector failures.
	internalContext = NewContext("github.com/remerge/cue")
	internalLogger  = NewLogger("github.com/remerge/cue")

	// Sending represents the number of sends currently in-process.
	// It is updated atomically and used to safely terminate workers.
	sending int32
)

// Collector is the interface representing event subscribers.  Log events are
// only generated and dispatched if collectors are registered with corresponding
// threshold levels.
//
// If a Collector implements the io.Closer interface, it's Close() method is
// called as part of termination.
type Collector interface {
	Collect(event *Event) error
}

// Logger is the interface for logging instances.
type Logger interface {
	EnabledFor(level Level) bool
	// WithFields returns a new logger instance with fields added to the current
	// logger's context.
	WithFields(fields Fields) Logger

	// WithValue returns a new logger instance with key and value added to the
	// current logger's context.
	WithValue(key string, value interface{}) Logger

	// Debug logs a message at the DEBUG level.
	Debug(message string)

	// Debugf logs a message at the DEBUG level using formatting rules from
	// the fmt package.
	Debugf(format string, values ...interface{})

	// Info logs a message at the INFO level.
	Info(message string)

	// Infof logs a message at the INFO level using formatting rules from
	// the fmt package.
	Infof(format string, values ...interface{})

	// Warn logs a message at the WARN level.
	Warn(message string)

	// Warnf logs a message at the WARN level using formatting rules from
	// the fmt package.
	Warnf(format string, values ...interface{})

	// Error logs the given error and message at the ERROR level and returns
	// the same error value. If err is nil, Error returns without emitting
	// a log event.
	Error(err error, message string) error

	// Errorf logs the given error at the ERROR level using formatting rules
	// from the fmt package and returns the same error value.  If err is nil,
	// Errorf returns without emitting a log event.
	Errorf(err error, format string, values ...interface{}) error

	// Panic logs the given cause and message at the FATAL level and then
	// calls panic(cause).  Panic does nothing is cause is nil.
	Panic(cause interface{}, message string)

	// Panicf logs the given cause at the FATAL level using formatting rules
	// from the fmt package and then calls panic(cause). Panicf does nothing if
	// cause is nil.
	Panicf(cause interface{}, format string, values ...interface{})

	// Recover recovers from panics and logs the recovered value and message
	// at the FATAL level.  Recover must be called via defer. If a logger's
	// Panic or Panicf method is used to trigger the panic, Recover returns
	// without emitting a new log event.  Recover does nothing if there's no
	// panic to recover.
	Recover(message string)

	// ReportRecovery logs the given cause and message at the FATAL level.
	// If used, it should be called from a deferred function after that
	// function has recovered from a panic.  In most cases, using the Recover
	// method directly is simpler.  However, sometimes it's necessary to test
	// whether a panic occurred or not.  In those cases, it's easier to use
	// the built-in recover() function and simply report the recovery via
	// ReportRecovery.  ReportRecovery does nothing if cause is nil.
	ReportRecovery(cause interface{}, message string)

	// Wrap returns a logging instance that skips one additional frame when
	// capturing frames for a call site.  Wrap should only be used when logging
	// calls are wrapped by an additional library function or method.
	Wrap() Logger
}

// logger is the default logger implementation
type logger struct {
	context    Context
	skipFrames int // Number of frames to skip when calling event.captureFrames.
}

// NewLogger returns a new logger instance using name for the context.
func NewLogger(name string) Logger {
	return &logger{
		context: NewContext(name),

		// We ensure that our logger.send* methods are called from a uniform
		// stack depth and that skipping 3 frames from those locations will
		// locate the original caller of our exported logging methods.
		skipFrames: 3,
	}
}

func (l *logger) String() string {
	return fmt.Sprintf("Logger(name=%s)", l.context.Name())
}

func (l *logger) WithFields(fields Fields) Logger {
	new := l.clone()
	new.context = new.context.WithFields(fields)
	return new
}

func (l *logger) WithValue(key string, value interface{}) Logger {
	new := l.clone()
	new.context = new.context.WithValue(key, value)
	return new
}

func (l *logger) Wrap() Logger {
	new := l.clone()
	new.skipFrames++
	return new
}

func (l *logger) Debug(message string) {
	l.send(DEBUG, nil, message)
}

func (l *logger) Debugf(format string, values ...interface{}) {
	l.sendf(DEBUG, nil, format, values...)
}

func (l *logger) Info(message string) {
	l.send(INFO, nil, message)
}

func (l *logger) Infof(format string, values ...interface{}) {
	l.sendf(INFO, nil, format, values...)
}

func (l *logger) Warn(message string) {
	l.send(WARN, nil, message)
}

func (l *logger) Warnf(format string, values ...interface{}) {
	l.sendf(WARN, nil, format, values...)
}

func (l *logger) Error(err error, message string) error {
	if err == nil {
		return nil
	}
	l.send(ERROR, err, message)
	return err
}

func (l *logger) Errorf(err error, format string, values ...interface{}) error {
	if err == nil {
		return nil
	}
	l.sendf(ERROR, err, format, values...)
	return err
}

func (l *logger) Panic(cause interface{}, message string) {
	if cause == nil {
		return
	}
	l.sendPanic(cause, message)
}

func (l *logger) Panicf(cause interface{}, format string, values ...interface{}) {
	if cause == nil {
		return
	}
	l.sendPanicf(cause, format, values...)
}

func (l *logger) Recover(message string) {
	cause := recover()
	if cause == nil || ourPanic() {
		return
	}
	l.sendRecovery(cause, message)
}

func (l *logger) ReportRecovery(cause interface{}, message string) {
	if cause == nil || ourPanic() {
		return
	}
	l.sendRecovery(cause, message)
}

func (l *logger) EnabledFor(level Level) bool {
	return level <= cfg.get().threshold
}

func (l *logger) send(level Level, err error, message string) {
	config := cfg.get()
	if level > config.threshold {
		return
	}

	event := newEvent(l.context, level, err, message)
	event.captureFrames(l.skipFrames, config.frames, config.errorFrames, false)
	l.dispatchEvent(event)
}

func (l *logger) sendf(level Level, err error, format string, values ...interface{}) {
	config := cfg.get()
	if level > config.threshold {
		return
	}

	event := newEventf(l.context, level, err, format, values...)
	event.captureFrames(l.skipFrames, config.frames, config.errorFrames, false)
	l.dispatchEvent(event)
}

func (l *logger) sendPanic(cause interface{}, message string) {
	config := cfg.get()
	if FATAL > config.threshold {
		doPanic(cause)
	}

	event := newEvent(l.context, FATAL, nil, message)
	err, ok := cause.(error)
	if !ok {
		err = errors.New(fmt.Sprint(cause))
	}
	event.Error = err
	event.captureFrames(l.skipFrames, config.frames, config.errorFrames, false)
	l.dispatchEvent(event)
	doPanic(cause)
}

func (l *logger) sendPanicf(cause interface{}, format string, values ...interface{}) {
	config := cfg.get()
	if FATAL > config.threshold {
		doPanic(cause)
	}

	event := newEventf(l.context, FATAL, nil, format, values...)
	err, ok := cause.(error)
	if !ok {
		err = errors.New(fmt.Sprint(cause))
	}
	event.Error = err
	event.captureFrames(l.skipFrames, config.frames, config.errorFrames, false)
	l.dispatchEvent(event)
	doPanic(cause)
}

func (l *logger) sendRecovery(cause interface{}, message string) {
	config := cfg.get()
	if FATAL > config.threshold {
		return
	}

	event := newEvent(l.context, FATAL, nil, message)
	err, ok := cause.(error)
	if !ok {
		err = errors.New(fmt.Sprint(cause))
	}
	event.Error = err
	event.captureFrames(l.skipFrames, config.frames, config.errorFrames, true)
	l.dispatchEvent(event)
}

func (l *logger) dispatchEvent(event *Event) {
	atomic.AddInt32(&sending, 1)
	defer atomic.AddInt32(&sending, -1)
	for _, entry := range cfg.get().registry {
		if entry.threshold >= event.Level && !entry.degraded {
			entry.worker.Send(event)
		}
	}
}

func (l *logger) clone() *logger {
	return &logger{
		context:    l.context,
		skipFrames: l.skipFrames,
	}
}

// Collect registers a Collector for the given threshold using synchronous
// event collection.  Any event logged within the specified threshold will
// be sent to the provided collector.  Thus a collector registered for the
// WARN level will receive WARN, ERROR, and FATAL events.
//
// All events sent to the collector are fully synchronous and block until
// the collector's Collect method returns successfully.  This is dangerous
// if the collector performs blocking operations or returns errors.
func Collect(threshold Level, c Collector) {
	collect(threshold, 0, c)
}

// CollectAsync registers a Collector for the given threshold using
// asynchronous event collection.  Any event logged within the specified
// threshold will be sent to the provided collector.  Thus a collector
// registered for the WARN level will receive WARN, ERROR, and FATAL events.
//
// CollectAsync registers asynchronous collectors.  It creates a buffered
// channel for the collector and starts a worker goroutine to service events.
// Logging calls return after queuing events to the collector channel.  If the
// channel's buffer is full, the event is dropped and a drop counter is
// incremented atomically.  This ensures asynchronous logging calls never
// block.  The worker goroutine detects changes in the atomic drop counter and
// surfaces drop events as collector errors.  See the cue/collector docs for
// details on collector error handling.
//
// When asynchronous logging is enabled, Close must be called to flush queued
// events on program termination.  Close is safe to call even if asynchronous
// logging isn't enabled -- it returns immediately if no events are queued.
// Note that ctrl+c and kill <pid> terminate Go programs without triggering
// cleanup code.  When using asynchronous logging, it's a good idea to register
// signal handlers to capture SIGINT (ctrl+c) and SIGTERM (kill <pid>).  See
// the Signals example and os/signals package docs for details.
func CollectAsync(threshold Level, bufsize int, c Collector) {
	collect(threshold, bufsize, c)
}

func collect(threshold Level, bufsize int, c Collector) {
	if c == nil {
		return
	}

	cfg.lock()
	defer cfg.unlock()

	new := cfg.get().clone()
	_, present := new.registry[c]
	if present {
		return
	}

	new.registry[c] = &entry{
		threshold: threshold,
		worker:    newWorker(c, bufsize),
	}
	new.updateThreshold()
	cfg.set(new)
}

// SetLevel changes a registered collector's threshold level.  The OFF value
// may be used to disable event collection entirely.  SetLevel may be called
// any number of times during program execution to dynamically alter collector
// thresholds.
func SetLevel(threshold Level, c Collector) {
	cfg.lock()
	defer cfg.unlock()

	new := cfg.get().clone()
	entry, present := new.registry[c]
	if !present {
		return
	}
	entry.threshold = threshold
	new.updateThreshold()
	cfg.set(new)
}

// SetFrames specifies the number of stack frames to collect for log events.
// The frames parameter specifies the frame count to collect for DEBUG, INFO,
// and WARN events.  The errorFrames parameter specifies the frame count to
// collect for ERROR and FATAL events.  By default, a single frame is collected
// for events of any level that matches a subscribed collector's threshold.
// This ensures collectors may log the file name, package, function, and line
// number of the logging call site for any collected event.  SetFrames may be
// used to alter this frame count, or disable frame collection entirely by
// specifying a 0 value for either parameter.  SetFrames may be called any
// number of times during program execution to dynamically alter frame
// collection.
//
// When using error reporting services, SetFrames should be called to increase
// the errorFrames parameter from the default value of 1 to a value that
// provides enough stack context to successfully diagnose reported errors.
func SetFrames(frames int, errorFrames int) {
	cfg.lock()
	defer cfg.unlock()

	new := cfg.get().clone()
	new.frames = frames
	new.errorFrames = errorFrames
	cfg.set(new)
}

// setDegraded is called by worker instances to temporarily disable a degraded
// collector
func setDegraded(c Collector, degraded bool) {
	cfg.lock()
	defer cfg.unlock()

	new := cfg.get().clone()
	entry, present := new.registry[c]
	if !present {
		return
	}
	entry.degraded = degraded
	new.updateThreshold()
	cfg.set(new)
}

// Close is used to terminate and flush asynchronous logging buffers.  Close
// signals each worker to silently drop new events, flush existing  buffered
// events, and then terminate worker goroutines.  If all events flush within
// the given timeout, Close returns nil.  Otherwise it returns an error.
// Close may be called regardless of asynchronous logging state. It returns
// immediately if no events are buffered and no workers need to be terminated.
//
// Close may be called multiple times throughout program execution.  If Close
// returns nil, cue is guaranteed to be reset to it's initial state.  This is
// useful for testing.
func Close(timeout time.Duration) error {
	result := make(chan error, 1)
	go terminateAsync(result)

	select {
	case err := <-result:
		return err
	case <-time.After(timeout):
		return errors.New("cue: timeout waiting for buffers to flush")
	}
}

func terminateAsync(result chan<- error) {
	cfg.lock()
	defer cfg.unlock()

	current := cfg.get()
	cfg.set(newConfig())

	terminateWorkers(current.registry)
	result <- nil
}

func terminateWorkers(reg registry) {
	// We have to wait until in-process sends are complete before signaling the
	// workers to terminate.  Otherwise, in-process sends could attempt sending
	// on a closed channel, which would panic.
	for atomic.LoadInt32(&sending) != 0 {
		runtime.Gosched() // Yield the processor
	}

	var wg sync.WaitGroup
	for _, entry := range reg {
		wg.Add(1)
		go func(worker worker) {
			flush := true
			worker.Terminate(flush)
			wg.Done()
		}(entry.worker)
	}
	wg.Wait()
}

// dispose terminates the collector, discards any buffered messages for it, and
// removes the collector from the registry entirely.
func dispose(c Collector) {
	cfg.lock()
	defer cfg.unlock()

	new := cfg.get().clone()
	entry, present := new.registry[c]
	if !present {
		return
	}

	delete(new.registry, c)
	new.updateThreshold()
	cfg.set(new)

	flush := false
	entry.worker.Terminate(flush)
}
