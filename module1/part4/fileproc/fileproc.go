package fileproc

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"sync"
)

var fileOpener = func(name string) (io.ReadCloser, error) {
	return os.Open(name)
}

type workerResult struct {
	lines int
	err   error
}

// ProcessFiles обрабатывает файлы параллельно в workers воркерах.
// Значения workers <= 0 нормализуются к 1.
// Пустой paths возвращает (0, nil) без запуска воркеров.
// Если хотя бы один файл не удалось обработать — возвращает объединённую ошибку
// (см. errors.Join), но обрабатывает все файлы до конца.
func ProcessFiles(paths []string, workers int) (totalLines int, err error) {
	if workers <= 0 {
		workers = 1
	}

	if len(paths) == 0 {
		return 0, nil
	}

	tasks := make(chan string, len(paths))
	results := make(chan workerResult, len(paths))

	var wg sync.WaitGroup

	workersCount := workers
	if workersCount > len(paths) {
		workersCount = len(paths)
	}

	wg.Add(workersCount)
	for i := 0; i < workersCount; i++ {
		go worker(tasks, results, &wg)
	}

	for _, path := range paths {
		tasks <- path
	}
	close(tasks)

	go func() {
		wg.Wait()
		close(results)
	}()

	var errs []error
	for res := range results {
		if res.err != nil {
			errs = append(errs, res.err)
		}
		totalLines += res.lines
	}

	return totalLines, errors.Join(errs...)
}

// CountFileLines открывает файл по указанному пути и считает количество строк в нем.
//
// Для чтения используется [bufio.Scanner]. Построчный обход продолжается до конца файла,
// либо до первой внутренней ошибки сканера.
//
// Функция гарантирует закрытие файла (освобождение дескриптора ОС) даже в случае,
// если в процессе чтения возникнет паника (panic). Если при закрытии файла
// произойдет ошибка, она будет корректно добавлена в возвращаемое значение err.
func CountFileLines(path string) (lines int, err error) {
	f, err := fileOpener(path)
	if err != nil {
		return 0, err
	}

	// Корректный return err даже при closeErr
	defer func() {
		closeErr := f.Close()
		if closeErr != nil && err == nil {
			err = closeErr
		} else if closeErr != nil {
			err = fmt.Errorf("%w (close: %w)", err, closeErr)
		}
	}()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		lines++
	}

	if scannerErr := scanner.Err(); scannerErr != nil {
		err = scannerErr
	}

	return
}

// Worker обрабатывает задачи из канала tasks.
// Перехватывает панику, чтобы не уронить всю программу.
func worker(tasks <-chan string, results chan<- workerResult, wg *sync.WaitGroup) {
	defer wg.Done()

	for path := range tasks {
		var res workerResult
		func() {
			defer func() {
				if r := recover(); r != nil {
					res = workerResult{
						err: fmt.Errorf("panic in worker for %s: %v", path, r),
					}
				}
			}()
			res.lines, res.err = CountFileLines(path)
		}()
		results <- res
	}
}
