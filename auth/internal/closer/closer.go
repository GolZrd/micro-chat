package closer

import (
	"log"
	"os"
	"os/signal"
	"sync"
)

// Глобальный приватный объект, инициализируем с помощью new()
var globalCloser = New()

// Эти 3 метода по сути внешние имплементации Closer
func Add(f ...func() error) {
	globalCloser.Add(f...)
}

func Wait() {
	globalCloser.Wait()
}

// Вызов всех функций закрытия
func CloseAll() {
	globalCloser.CloseAll()
}

type Closer struct {
	mu    sync.Mutex
	once  sync.Once
	done  chan struct{}
	funcs []func() error
}

// New возвращает новый экземпляр Closer
func New(sig ...os.Signal) *Closer {
	c := &Closer{done: make(chan struct{})}
	if len(sig) > 0 {
		go func() {
			ch := make(chan os.Signal, 1)
			signal.Notify(ch, sig...)
			<-ch
			signal.Stop(ch)
			c.CloseAll()
		}()
	}
	return c
}

// Add добавляет ряд функций в слайс funcs
func (c *Closer) Add(f ...func() error) {
	c.mu.Lock()
	c.funcs = append(c.funcs, f...)
	c.mu.Unlock()
}

// Wait блокируется пока что то не прилетит в done. То есть пока все закрытия не произошли.
func (c *Closer) Wait() {
	<-c.done
}

// CloseAll вызывает все функции закрытия
func (c *Closer) CloseAll() {
	c.once.Do(func() {
		defer close(c.done)

		c.mu.Lock()
		funcs := c.funcs
		c.funcs = nil
		c.mu.Unlock()

		// вызываем все функции закрытия асинхронно
		errs := make(chan error, len(funcs))
		for _, f := range funcs {
			go func(f func() error) {
				errs <- f()
			}(f)
		}

		for i := 0; i < cap(errs); i++ {
			if err := <-errs; err != nil {
				log.Println("error returned from closer:", err)
			}
		}
	})
}
