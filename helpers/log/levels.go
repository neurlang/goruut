package log

import "github.com/sirupsen/logrus"

// Trace

func Trace0(err error) {
	if err != nil {
		logrus.Traceln(err)
	}
}

func Trace1[T any](t T, err error) T {
	if err != nil {
		logrus.Traceln(err)
	}
	return t
}
func Trace2[S, T any](s S, t T, err error) (S, T) {
	if err != nil {
		logrus.Traceln(err)
	}
	return s, t
}

// Debug

func Debug0(err error) {
	if err != nil {
		logrus.Debugln(err)
	}
}

func Debug1[T any](t T, err error) T {
	if err != nil {
		logrus.Debugln(err)
	}
	return t
}
func Debug2[S, T any](s S, t T, err error) (S, T) {
	if err != nil {
		logrus.Debugln(err)
	}
	return s, t
}

// Info

func Info0(err error) {
	if err != nil {
		logrus.Infoln(err)
	}
}

func Info1[T any](t T, err error) T {
	if err != nil {
		logrus.Infoln(err)
	}
	return t
}
func Info2[S, T any](s S, t T, err error) (S, T) {
	if err != nil {
		logrus.Infoln(err)
	}
	return s, t
}

// Warning

func Warn0(err error) {
	if err != nil {
		logrus.Warnln(err)
	}
}

func Warn1[T any](t T, err error) T {
	if err != nil {
		logrus.Warnln(err)
	}
	return t
}
func Warn2[S, T any](s S, t T, err error) (S, T) {
	if err != nil {
		logrus.Warnln(err)
	}
	return s, t
}

// Error

func Error0(err error) {
	if err != nil {
		logrus.Errorln(err)
	}
}

func Error1[T any](t T, err error) T {
	if err != nil {
		logrus.Errorln(err)
	}
	return t
}
func Error2[S, T any](s S, t T, err error) (S, T) {
	if err != nil {
		logrus.Errorln(err)
	}
	return s, t
}
func Error3[S, T, U any](s S, t T, u U, err error) (S, T, U) {
	if err != nil {
		logrus.Errorln(err)
	}
	return s, t, u
}

// Fatal

func Fatal0(err error) {
	if err != nil {
		logrus.Fatalln(err)
	}
}

func Fatal1[T any](t T, err error) T {
	if err != nil {
		logrus.Fatalln(err)
	}
	return t
}
func Fatal2[S, T any](s S, t T, err error) (S, T) {
	if err != nil {
		logrus.Fatalln(err)
	}
	return s, t
}

// Error

func Panic0(err error) {
	if err != nil {
		logrus.Panicln(err)
	}
}

func Panic1[T any](t T, err error) T {
	if err != nil {
		logrus.Panicln(err)
	}
	return t
}
func Panic2[S, T any](s S, t T, err error) (S, T) {
	if err != nil {
		logrus.Panicln(err)
	}
	return s, t
}

// Exception

func Exception0(err error) {
	if err != nil {
		logrus.Errorln(err)
		panic(err.Error())
	}
}

func Exception1[T any](t T, err error) T {
	if err != nil {
		logrus.Errorln(err)
		panic(err.Error())
	}
	return t
}
