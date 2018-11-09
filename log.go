package main

import (
    "bufio"
    "fmt"
    "io"
    "os"
    "time"
    "strings"
    "regexp"
    //redis 
    "github.com/mediocregopher/radix.v2/redis"
    "log"
)

// Struct que guarda os logs
type Logs struct {
    Data  string
    Ip string
    Comando  string
    Username  string
    Local string
}

func criaLogs(reply map[string]string) (*Logs, error) {
    linhaLog := new(Logs)
    linhaLog.Data = reply["data"]
    //linhaLog.Ip = reply["ip"]
    linhaLog.Comando = reply["comando"]
    linhaLog.Username = reply["user"]
    linhaLog.Local = reply["local"]

    return linhaLog, nil
}

func findData(line string) string {
    var re = regexp.MustCompile(`[a-zA-Z]+  \d{1} \d{2}:\d{2}:\d{2}`) 
    if len(re.FindStringIndex(line)) > 0 {
        return re.FindString(line)
    } else {
        return ""
    }
}

func findUsername(line string) string {
    var re = regexp.MustCompile(`USER=[a-zA-Z]+ ;`) 
    if len(re.FindStringIndex(line)) > 0 {
        return re.FindString(line)
    } else {
        if len(findBetween(string(line), "nvalid user ", " from")) > 0 {
            return findBetween(string(line), "nvalid user ", " from")
        }
    }
    return ""
}

func findLocal(line string) string {
    var re = regexp.MustCompile(`PWD=[a-zA-Z0-9!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?]+ ;`) 
    if len(re.FindStringIndex(line)) > 0 {
        return re.FindString(line)
    } else {
        return ""
    }
}

func findIP(input string) string {
    numBlock := "(25[0-5]|2[0-4][0-9]|1[0-9][0-9]|[1-9]?[0-9])"
    regexPattern := numBlock + "\\." + numBlock + "\\." + numBlock + "\\." + numBlock
    regEx := regexp.MustCompile(regexPattern)
    return regEx.FindString(input)
}

func findCommand(value string, a string) string {
    pos := strings.LastIndex(value, a)
    if pos == -1 {
        return ""
    }
    adjustedPos := pos + len(a)
    if adjustedPos >= len(value) {
        return ""
    }
    return value[adjustedPos:len(value)]
}

func findBetween(value string, a string, b string) string {
    // Get substring between two strings.
    posFirst := strings.Index(value, a)
    if posFirst == -1 {
        return ""
    }
    posLast := strings.Index(value, b)
    if posLast == -1 {
        return ""
    }
    posFirstAdjusted := posFirst + len(a)
    if posFirstAdjusted >= posLast {
        return ""
    }
    return value[posFirstAdjusted:posLast]
}

func tail(filename string, out io.Writer) {
    //INICIA CONEXÃO COM REDIS
     conn, err := redis.Dial("tcp", "localhost:6379")
    if err != nil {
        log.Fatal(err)
    }
    defer conn.Close()
    //ABRE ARQUIVO LOG
    f, err := os.Open(filename)
    if err != nil {
        panic(err)
    }
    defer f.Close()
    r := bufio.NewReader(f)
    info, err := f.Stat()
    if err != nil {
        panic(err)
    }
    oldSize := info.Size()
    for {
        i := 0
        for line, prefix, err := r.ReadLine(); err != io.EOF; line, prefix, err = r.ReadLine() {
            if prefix {
            } else {
                comando := findCommand(string(line), "COMMAND=")
                auth := findBetween(string(line), "nvalid user ", " from")
                if len(comando) != 0 || len(auth) !=0 {
                    fmt.Println("\n####### NOVA INSERÇÃO #######")
                    data := findData(string(line))
                    fmt.Println("Data: ",data)
                    username := findUsername(string(line))
                    fmt.Println("User: ",username)
                    ip := findIP(string(line))
                    fmt.Println("IP: ", ip)
                    local := ""
                    if len(auth) > 0 {
                        comando = findBetween(string(line), "]: ", "user") + "user"
                    }else {
                        local := findLocal(string(line))
                        fmt.Println("Local: ",local)
                    }
                    fmt.Println("Comando: ", comando)
                    i++ // id da linha a ser adicionada
                    id := fmt.Sprintf("%s%d", "logs:", i)
                    //INSERE NO BD
                    resp := conn.Cmd("HMSET", id, "data", data, "ip", ip, "comando",comando, "user", username, "local", local)
                    if resp.Err != nil {
                        log.Fatal(resp.Err)
                    } else {
                        fmt.Println("Adicionado.")
                    }
                    //Pega os dados
                    reply, err := conn.Cmd("HGETALL", id).Map()
                    if err != nil {
                        log.Fatal(err)
                    }
                    //cria e imprime nova linha de log do map[string]string.
                    linhaLog, err := criaLogs(reply)
                    if err != nil {
                        log.Fatal(err)
                    }
                    fmt.Println("IMPRIMINDO: ", linhaLog)
                }
            }
        }
        pos, err := f.Seek(0, io.SeekCurrent)
        if err != nil {
            panic(err)
        }
        for {
            time.Sleep(time.Second)
            newinfo, err := f.Stat()
            if err != nil {
                panic(err)
            }
            newSize := newinfo.Size()
            if newSize != oldSize {
                if newSize < oldSize {
                    f.Seek(0, 0)
                } else {
                    f.Seek(pos, io.SeekStart)
                }
                r = bufio.NewReader(f)
                oldSize = newSize
                break
            }
        }
    }
}

func main() {
    tail("/home/alpaca/Documents/NCC/goLogs/auth.log", os.Stdout)
}