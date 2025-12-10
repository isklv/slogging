# Introduction #
## Overview | Research ##
Пакет для распределенного логирования и трассировки с помощью встроенного пакета slog в Graylog с поддержкой всех используемых средств связи (AMQP, gRPC, HTTP). Под поддержкой имеется ввиду реализацию возможности передачи X-B3 заголовков с помощью встроенных методов со стороны сервера, а также реализацию чтения этих же заголовков со стороны клиента, и дальнейшей передаче их в контекст и создания экземпляра логера с дополнительными X-B3 полями.  

Также в процессе разработки было пересмотрено отношение к X-B3 заголовкам, и так как у нас нет полноценной системы централизованной трассировки, а эту роль выполняет Graylog было принято решение избавиться от всех X-B3 заголовков кроме X-B3-TraceId за ненадобностью и бесполезностью прочих. 

Изучение X-B3 трассировки по Zipkin дало понять, что до этого мы неправильно генерировали X-B3-TraceId, в оригинале это должно быть 64 или 128 битное шестнадцатеричное число, у нас же в некоторых местах использовался UUID (исправлено). (https://github.com/openzipkin/b3-propagation)

Был пересмотрен и подробно изучен вопрос передачи логгера для комфортной и быстрой работы с ним, решением стало передача логгера в контексте, выполняющем утилитарную роль.
В следующей статье подробно поднят этот вопрос, ознакомтесь если возникнут вопросы
(https://www.kaznacheev.me/posts/en/where-to-place-logger-in-golang/)  
Теперь необходима передача контекста на всем пути работы программы, несмотря на то, что у нас это не всегда использовалось, это правильный подход.

Самостоятельная реализация и актуализация отправки по прикладному протоколу GELF данных в Graylog позволила уменьшить размер payload, путем устранения deprecated GELF data, которая потеряла актуальность в связи с обновлением протокола GELF и технологии Graylog, соответственно повысилась производительность, помимо этого сам пакет для логирования slog выйгрывает перед логрусом в 10 раз по скорости. (https://github.com/betterstack-community/go-logging-benchmarks?tab=readme-ov-file)

Был реализован утилитарный функционал для пакета slog, позволяющий создать fanout handler, так как slog не умел это из коробки, так же присутствует возможность гибкой настройки логгера как в стандартный вывод, так и в Graylog. Также был затронут UDP Batching, честно говоря я хотел переписать его с нуля, но когда побольше разобрался в этой теме, понял что он и так неплохо реализован, для этого был использован вспомогательный пакет slog-graylog. (https://github.com/samber/slog-graylog)  

## Сопоставление уровней пакета и Graylog ##
| slog  | Graylog |
|-------|---------|
| Debug | 7       |
| Info  | 6       |
| Warn  | 4       |
| Error | 3       |
| Fatal | 2       |



Был рассмотрен вопрос добавления уровня Fatal с os.Exit(1), от реализации пришлось отказаться, как я считаю всех представленных уровней достаточно, а остальные лишь излишнее усложнение, к тому же разработчики пакета slog сами отказались от добавления этого уровня, и предложили писать os.Exit(1) там где это нужно (потерял ссылку на github issue). К тому же для реализации кастомного уровня необходима еще одна инкапсуляция из-за чего код может стать слишком сложным и поддерживаемым. В данный момент реализация представлена без лишней инкапсуляции.

# Usage #

## Getting ##
```bash
go env -w GONOPROXY=github.com/isklv/*
go get github.com/isklv/slogging@v1.0.7
```

## Initialization ##
```bash
opts := slogging.NewOptions().InGraylog("localhost:12201", "application_name")
sl := slogging.NewLogger(opts)
```
Важно: в примере представлены все возможные опции, если ничего не указывать выставятся стандартные
## Описание опций настройки ##
|                     | Установлена                                                                                                                                                                                                                        | Не установлена                                                                           |
|---------------------|------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|------------------------------------------------------------------------------------------|
| InGraylog(params)   | Отправка логов в Graylog по протоколу UDP по указанному адресу,  учитывая выставленный уровень логов, все логи отправляемые в Graylog будут иметь указанное дополнительное поле "container_name"                                   | Отправка логов в Graylog отсутствует                                                     |
| SetLevel(params)    | Установка указанного уровня логов для записи в стандартный поток вывода                                                                                                                                                            | Будет установлен стандартный уровень "info" для записи в стандартный поток вывода        |
| WithSource(params)  | Добавление к логам для записи в стандартный поток вывода полей, отвечающих за источник лога (функция, файл, строка)                                                                                                                | Дополнительные поля, отвечающие за источник лога отсутствуют в стандартном потоке вывода |
| SetDefault(bool)    | В случае true, логгер будет установлен стандартным, то есть при вызове slog.Info например из любой точки программы, будет использован инициализированный логгер, если была включена отправка в Graylog, она тоже будет произведена | В случае false будет использован стандартный логгер slog.Default при slog.Info например  |

---
# Пример HTTP Middleware (Принимаем trace заголовок из HTTP запроса) #
## Gin ##
```
import (
    "github.com/gin-gonic/gin"
	ginsl "github.com/isklv/slogging/http/gin"
)

func main() {
	opts := slogging.NewOptions().InGraylog("localhost:12201", "application_name")
	sl := slogging.NewLogger(opts)

    // Можем передать текущий сконфигурированный логгер
	traceMW := ginsl.TraceMiddleware(sl.Logger)
	
	// Можем сконфигурировать и передать новый экземпляр логгера
	// traceMW := ginsl.TraceMiddleware(sl.With("module", "gin-http"))
	
	// А можем дефолтный slog.Default(), актуально если доп конфигурация не требуется и при slogging.SetDefault(true)
	// traceMW := ginsl.TraceMiddleware(slog.Default())

	r := gin.New()

	exGroup := r.Group("/example")
	exGroup.Use(traceMW)

	exGroup.POST("/action", actionHandler)
}

func actionHandler(c *gin.Context) {
    // Теперь тут в c.Request.Context() контексте лежат trace заголовки ))
	ctx := c.Request.Context()

	slogging.L(ctx).Info("hello world =)")
}
```

## Mux ##
```
import (
    "github.com/gorilla/mux"
	muxsl "github.com/isklv/slogging/http/mux"
)


func main() {
	opts := slogging.NewOptions().InGraylog("localhost:12201", "application_name")
	sl := slogging.NewLogger(opts)

    // Можем передать текущий сконфигурированный логгер
	traceMW := muxsl.TraceMiddleware(sl.Logger)
	
	// Можем сконфигурировать и передать новый экземпляр логгера
	//traceMW := muxsl.TraceMiddleware(sl.With("module", "mux-http"))
	
	// А можем дефолтный slog.Default(), актуально если доп конфигурация не требуется и при slogging.SetDefault(true)
	//traceMW := muxsl.TraceMiddleware(slog.Default())

	r := mux.NewRouter()

	exGroup := r.PathPrefix("/example").Subrouter()
	exGroup.Use(traceMW)

	exGroup.HandleFunc("/action", actionHandler)
}

func actionHandler(w http.ResponseWriter, r *http.Request) {
    // Теперь тут в r.Context() контексте лежат trace заголовки ))
	ctx := r.Context()


	slogging.L(ctx).Info("hello world =)")
}
```

## Native ##
```
import (
    "net/http"
	sl "github.com/isklv/slogging/http"
)

func main() {
	opts := slogging.NewOptions().InGraylog("localhost:12201", "application_name")
	sl := slogging.NewLogger(opts)

    // Можем передать текущий сконфигурированный логгер
	traceMW := sl.TraceMiddleware(l.Logger)
	
	// Можем сконфигурировать и передать новый экземпляр логгера
	// traceMW := sl.TraceMiddleware(l.With("module", "http"))
	
	// А можем дефолтный slog.Default(), актуально если доп конфигурация не требуется и при slogging.SetDefault(true)
	// traceMW := sl.TraceMiddleware(slog.Default())

	http.HandleFunc("/example/action", traceMW(actionHandler))
}

func actionHandler(w http.ResponseWriter, r *http.Request) {
    // Теперь тут в r.Context() контексте лежат trace заголовки ))
	ctx := r.Context()

	slogging.L(ctx).Info("hello world =)")
}
```
---

# Пример GRPC Interceptor (Принимаем trace заголовок из GRPC сообщения) #
```
import (
    grpcsl "github.com/isklv/slogging/grpc"
	"google.golang.org/grpc"
)

type server struct {
	pb.UnimplementedGreeterServer
}

func (s *server) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloResponse, error) {
	// Тут в контексте будет лежать trace заголовок
	slogging.L(ctx).Info("hello world =')")

	return &pb.HelloResponse{Greeting: "Hello " + in.GetName()}, nil
}

func main() {
	opts := slogging.NewOptions().InGraylog("localhost:12201", "application_name")
	sl := slogging.NewLogger(opts)

	lis, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

    // Можем передать текущий сконфигурированный логгер
	traceIC := grpcsl.TraceInterceptor(l.Logger)
	
	// Можем сконфигурировать и передать новый экземпляр логгера
	//traceIC := grpcsl.TraceInterceptor(l.With("module", "grpc"))
	
	// А можем дефолтный slog.Default(), актуально если доп конфигурация не требуется и при slogging.SetDefault(true)
	//traceIC := grpcsl.TraceInterceptor(slog.Default())

	s := grpc.NewServer(
		// Тут пробрасываем наш interceptor в сервер
		grpc.UnaryInterceptor(traceIC))
	pb.RegisterGreeterServer(s, &server{})

	log.Printf("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
```
---
# Пример отправки trace заголовка по HTTP #
```
import (
    sl "github.com/isklv/slogging/http"
	"net/http"
)

// В контексте должно лежать trace поле
func sendHTTP(ctx context.Context) {
	req, _ := http.NewRequest("GET", "/", nil)
	req := sl.TraceRequest(ctx, req)

    // Отправится запрос где в заголовках будет trace заголовок
	http.DefaultClient.Do(req)
}
```
---
# Пример отправки trace метадаты по GRPC #
```
import (
    grpcsl "github.com/isklv/slogging/grpc"
)

// В контексте должно лежать trace поле
func sendGRPC(ctx context.Context, c pb.GreeterClient) {
	ctx = grpcsl.TraceMetadata(ctx)

	// Отправится trace заголовок в метадате
	c.SayHello(ctx, &pb.HelloRequest{Name: defaultName})
}
```
---



## Логирование из контекста с логгером ##
В примере приведены также экстра поля для логирования в Graylog
```
slogging.L(ctx).Info("example log message",
		slogging.ErrAttr(errors.New("example error message")),
		slogging.StringAttr("hello", "world"),
		slogging.IntAttr("bye", 12),
		slogging.AnyAttr("data", object),
		slogging.FloatAttr("bye", 14.88),
		slogging.TimeAttr("timestamp", time.Now()),
	)
```
---
## Использование логгера для вывода Request и Response ##
```
client := &http.Client{}
 
    reqBody, _ := json.Marshal(map[string]string{
        "title":  "foo",
        "body":   "bar",
        "userId": "1",
    })
 
    req, err := http.NewRequest("POST", "https://jsonplaceholder.typicode.com/posts", bytes.NewBuffer(reqBody))
    if err != nil {
        log.Fatalf("Error creating request: %v", err)
    }
 
    req.Header.Set("Content-Type", "application/json")
 
    // Контекст с traceId
    ctx := context.WithValue(context.Background(), "traceId", "12345")
 
    // Логируем исходящий запрос
    L(ctx).Info("Outgoing Request", slogging.RequestAttr(req)...)
 
    start := time.Now()
    resp, err := client.Do(req)
    if err != nil {
        L(ctx).Error("Request Failed", slog.String("error", err.Error()))
        return
    }
    defer resp.Body.Close()
 
    // Логируем входящий ответ
    L(ctx).Info("Incoming Response", slogging.ResponseAttr(resp, time.Since(start))...)
}
```
---
## Создание логгера с новыми полями для передачи дальше ##
```
log := slogging.L(ctx).With(
		slogging.StringAttr("module", "keycloak"))

ctx = slogging.ContextWithLogger(ctx, log)
// Передаем контекст дальше, TraceID в нем сохранится, в логгере добавится дополнительное поле
```
---
## Создание контекста с логгером со случайным TraceID (полезно для тестов) ##
```
ctx := slogging.Context()
```
---

## Функция trace exemplar для Prometheus ##
```
import (
    "github.com/isklv/slogging/prometheus"
)

func main() {
    traceFunc := prometheus.TraceExemplar
}

```
