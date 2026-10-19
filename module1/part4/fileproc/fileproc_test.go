package fileproc

import (
	"errors"
	"io"
	"strings"
	"sync/atomic"
	"testing"
)

type mockFile struct {
	io.Reader
	closedCount *int32
	panicOnRead bool
}

func (m *mockFile) Close() error {
	if m.closedCount != nil {
		atomic.AddInt32(m.closedCount, 1)
	}
	return nil
}

func (m *mockFile) Read(p []byte) (n int, err error) {
	if m.panicOnRead {
		panic("simulated panic during scan")
	}

	return m.Reader.Read(p)
}

func TestProcessFiles_OK(t *testing.T) {
	var closedFiles int32

	mockFiles := map[string]string{
		"file1.txt": "row1\nrow2\n",
		"file2.txt": "row1\n",
		"file3.txt": "row1\nrow2\nrow3\n",
	}

	oldOpener := fileOpener
	t.Cleanup(func() { fileOpener = oldOpener })

	fileOpener = func(name string) (io.ReadCloser, error) {
		content, ok := mockFiles[name]
		if !ok {
			return nil, errors.New("File not found")
		}
		return &mockFile{
			Reader:      strings.NewReader(content),
			closedCount: &closedFiles,
		}, nil
	}

	total, err := ProcessFiles([]string{"file1.txt", "file2.txt", "file3.txt"}, 2)
	if err != nil {
		t.Fatalf("Ожидался успешный запуск, получена ошибка: %v", err)
	}

	expectedLines := 6 // 2 + 1 + 3
	if total != expectedLines {
		t.Errorf("Ожидалось строк: %d, получено: %d", expectedLines, total)
	}

	// Проверяем, что все 3 открытых файла были закрыты через defer
	if atomic.LoadInt32(&closedFiles) != 3 {
		t.Errorf("Ожидалось закрытие 3 файлов, закрыто: %d", closedFiles)
	}
}

func TestProcessFiles_FileNotFound(t *testing.T) {
	var closedFiles int32

	mockFiles := map[string]string{
		"exist1.txt": "row1\n",
		"exist2.txt": "row1\nrow2\n",
	}

	oldOpener := fileOpener
	t.Cleanup(func() { fileOpener = oldOpener })

	fileOpener = func(name string) (io.ReadCloser, error) {
		content, ok := mockFiles[name]
		if !ok {
			return nil, errors.New("file not found")
		}
		return &mockFile{
			Reader:      strings.NewReader(content),
			closedCount: &closedFiles,
		}, nil
	}

	// Пытаемся обработать 3 файла, missing.txt не существует
	_, err := ProcessFiles([]string{"exist1.txt", "missing.txt", "exist2.txt"}, 2)
	if err == nil {
		t.Error("Ожидалась ошибка FileNotFound, но метод вернул err = nil")
	}

	// Должно быть успешно закрыто ровно 2 файла, которые смогли открыться
	if atomic.LoadInt32(&closedFiles) != 2 {
		t.Errorf("Должно быть закрыто 2 файла, закрыто: %d", closedFiles)
	}
}

func TestProcessFiles_Panic(t *testing.T) {
	var closedFiles int32

	oldOpener := fileOpener
	t.Cleanup(func() { fileOpener = oldOpener })

	fileOpener = func(name string) (io.ReadCloser, error) {
		return &mockFile{
			Reader:      strings.NewReader("some data\n"),
			closedCount: &closedFiles,
			panicOnRead: true, // Провоцирует панику при попытке чтения (внутри bufio.Scanner)
		}, nil
	}

	// Запускаем. Если recover в worker() не отработает — тест упадет с крашем всей программы
	_, err := ProcessFiles([]string{"panic_file.txt"}, 1)

	if err == nil {
		t.Error("Ожидалась ошибка из-за перехваченной паники, но err = nil")
	}

	if !strings.Contains(err.Error(), "panic in worker") {
		t.Errorf("Ожидался текст ошибки про панику, получено: %v", err)
	}

	// Файл должен быть закрыт дефером внутри CountFileLines, несмотря на панику в сканере
	if atomic.LoadInt32(&closedFiles) != 1 {
		t.Errorf("Файл должен был закрыться по defer при панике, но closedCount = %d", closedFiles)
	}
}

func TestProcessFiles_EdgeCases(t *testing.T) {
	var closedFiles int32
	// Подменяем fileOpener, чтобы тесты не зависели от файловой системы
	oldFileOpener := fileOpener
	t.Cleanup(func() { fileOpener = oldFileOpener })

	fileOpener = func(name string) (io.ReadCloser, error) {
		return &mockFile{
			Reader:      strings.NewReader("line1\nline2\nline3\n"),
			closedCount: &closedFiles,
		}, nil
	}

	tests := []struct {
		name           string
		paths          []string
		workers        int
		expectedLines  int
		exprectedClose int32
		expectErr      bool
	}{
		{
			name:           "workers = 0",
			paths:          []string{"file1.txt", "file2.txt"},
			workers:        0, // Должно автоматически сброситься на 1 worker
			expectedLines:  6,
			exprectedClose: 2,
			expectErr:      false,
		},
		{
			name:           "workers < 0",
			paths:          []string{"file1.txt"},
			workers:        -5, // Должно автоматически сброситься на 1 worker
			expectedLines:  3,
			exprectedClose: 1,
			expectErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			atomic.StoreInt32(&closedFiles, 0)
			lines, err := ProcessFiles(tt.paths, tt.workers)
			if (err != nil) != tt.expectErr {
				t.Errorf("ProcessFiles() error = %v, expectErr %v", err, tt.expectErr)
				return
			}
			if lines != tt.expectedLines {
				t.Errorf("ProcessFiles() lines = %v, expected %v", lines, tt.expectedLines)
			}

			if atomic.LoadInt32(&closedFiles) != tt.exprectedClose {
				t.Errorf("ProcessFiles() closedFiles = %d, expected %d", atomic.LoadInt32(&closedFiles), tt.exprectedClose)
			}
		})
	}
}

func TestProcessFiles_EmptyPaths(t *testing.T) {
	var closedFiles int32

	oldOpener := fileOpener
	t.Cleanup(func() { fileOpener = oldOpener })

	fileOpener = func(name string) (io.ReadCloser, error) {
		return &mockFile{
			Reader:      strings.NewReader(""),
			closedCount: &closedFiles,
		}, nil
	}

	total, err := ProcessFiles([]string{}, 4)
	if err != nil {
		t.Fatalf("Ожидался успешный запуск при пустом списке path, получена ошибка: %v", err)
	}

	if total != 0 {
		t.Errorf("Ожидалось 0 строк, получено: %d", total)
	}

	if atomic.LoadInt32(&closedFiles) != 0 {
		t.Errorf("Файлы не должны были открываться и закрываться, закрыто: %d", closedFiles)
	}
}

func TestCountFileLines_EmptyFile(t *testing.T) {
	var closedFiles int32

	oldOpener := fileOpener
	t.Cleanup(func() { fileOpener = oldOpener })

	fileOpener = func(name string) (io.ReadCloser, error) {
		if name != "empty.txt" {
			return nil, errors.New("File not found")
		}
		return &mockFile{
			Reader:      strings.NewReader(""), // Содержимое пустое
			closedCount: &closedFiles,
		}, nil
	}

	total, err := CountFileLines("empty.txt")
	if err != nil {
		t.Fatalf("Ожидался успешный запуск для пустого файла, получена ошибка: %v", err)
	}

	if total != 0 {
		t.Errorf("Ожидалось 0 строк для пустого файла, получено: %d", total)
	}

	if atomic.LoadInt32(&closedFiles) != 1 {
		t.Errorf("Ожидалось закрытие 1 пустого файла, закрыто: %d", closedFiles)
	}
}

func TestProcessFiles_MoreWorkersThanPaths(t *testing.T) {
	var closedFiles int32

	mockFiles := map[string]string{
		"file1.txt": "row1\n",
		"file2.txt": "row1\nrow2\n",
	}

	oldOpener := fileOpener
	t.Cleanup(func() { fileOpener = oldOpener })

	fileOpener = func(name string) (io.ReadCloser, error) {
		content, ok := mockFiles[name]
		if !ok {
			return nil, errors.New("File not found")
		}
		return &mockFile{
			Reader:      strings.NewReader(content),
			closedCount: &closedFiles,
		}, nil
	}

	// Передаем 10 воркеров на 2 файла (пути)
	total, err := ProcessFiles([]string{"file1.txt", "file2.txt"}, 10)
	if err != nil {
		t.Fatalf("Ожидался успешный запуск при workers > len(paths), получена ошибка: %v", err)
	}

	expectedLines := 3 // 1 + 2
	if total != expectedLines {
		t.Errorf("Ожидалось строк: %d, получено: %d", expectedLines, total)
	}

	if atomic.LoadInt32(&closedFiles) != 2 {
		t.Errorf("Ожидалось закрытие 2 файлов, закрыто: %d", closedFiles)
	}
}

var (
	errTestOpen1 = errors.New("mock open error 1")
	errTestOpen2 = errors.New("mock open error 2")
)

func TestProcessFiles_ErrorsJoin(t *testing.T) {
	var closedFiles int32

	oldOpener := fileOpener
	t.Cleanup(func() { fileOpener = oldOpener })

	// Имитируем ошибку открытия для конкретных файлов
	fileOpener = func(name string) (io.ReadCloser, error) {
		if name == "fail1.txt" {
			return nil, errTestOpen1
		}
		if name == "fail2.txt" {
			return nil, errTestOpen2
		}
		return &mockFile{
			Reader:      strings.NewReader("row1\n"),
			closedCount: &closedFiles,
		}, nil
	}

	paths := []string{"fail1.txt", "file1.txt", "fail2.txt"}
	total, err := ProcessFiles(paths, 3)

	if err == nil {
		t.Fatal("Ожидалась ошибка, получен nil")
	}

	// Идиоматичная проверка через errors.Is гарантирует, что тесты не сломаются
	// при изменении порядка горутин или формата вывода строк
	if !errors.Is(err, errTestOpen1) {
		t.Errorf("Ожидалось, что цепочка ошибок содержит errTestOpen1, но errors.Is вернул false")
	}
	if !errors.Is(err, errTestOpen2) {
		t.Errorf("Ожидалось, что цепочка ошибок содержит errTestOpen2, но errors.Is вернул false")
	}

	expectedLines := 1
	if total != expectedLines {
		t.Errorf("Ожидалось строк: %d, получено: %d", expectedLines, total)
	}

	if atomic.LoadInt32(&closedFiles) != 1 {
		t.Errorf("Ожидалось закрытие 1 успешного файла, закрыто: %d", closedFiles)
	}
}
