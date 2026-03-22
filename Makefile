BINARY_NAME=apartapatia-containers

.PHONY: all build run clean

all: build

build: 
	  go build -o $(BINARY_NAME) main.go
			
run: build
	  sudo ./$(BINARY_NAME) run

test: build
	  @echo "Тестирование запуска приложения"
	  sudo ./$(BINARY_NAME) run < /dev/null

clean:
	  rm -f $(BINARY_NAME)
