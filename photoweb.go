package main
import (
    "io"
    "log"
    "net/http"
    "os"
    "io/ioutil"
    "html/template"
    "path"
    "runtime/debug"
    "fmt"
)

const (
    ListDir = 0x0001
    UPLOAD_DIR = "./uploads"
    TEMPLATE_DIR = "./views"
)

var templates = make(map[string]*template.Template)
func init() {
    fileInfoArr, err := ioutil.ReadDir(TEMPLATE_DIR)
    check(err)
    var templateName, templatePath string
    for _, fileInfo := range fileInfoArr {
        templateName = fileInfo.Name()
        if ext := path.Ext(templateName); ext != ".html" {
            continue
        }
        templatePath = TEMPLATE_DIR + "/" + templateName
        log.Println("Loading template:", templatePath)
        t := template.Must(template.ParseFiles(templatePath))
        fmt.Println("filename is ",templateName)
        templates[templateName] = t
    }
}
func uploadHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method == "GET" {
        /*
        io.WriteString(w, "<form method=\"POST\" action=\"/upload\" "+
            " enctype=\"multipart/form-data\">"+
            "Choose an image to upload: <input name=\"image\" type=\"file\" />"+
            "<input type=\"submit\" value=\"Upload\" />"+
            "</form>")
            */
            /*
        t, err := template.ParseFiles("upload.html")
        if err != nil {
        http.Error(w, err.Error(),http.StatusInternalServerError)
        return
        }
        t.Execute(w, nil)   
        */
        if err := renderHtml(w, "upload.html", nil); err != nil{
            /*
            http.Error(w, err.Error(),
            http.StatusInternalServerError)
            return
            */
            check(err)
        }
        
    }
    if r.Method == "POST" {
        f, h, err := r.FormFile("image")
         /* if err != nil {
        
        http.Error(w, err.Error(),
        http.StatusInternalServerError)
        return
        
        
        }*/
        check(err)
        filename := h.Filename
        defer f.Close()
        t, err := os.Create(UPLOAD_DIR + "/" + filename)
         /* if err != nil {
        
        http.Error(w, err.Error(),
        http.StatusInternalServerError)
        return
        
        
        }*/
        check(err)
        defer t.Close()
        //_, err := io.Copy(t, f)//no new variables on left side of :=
       //check(err)
       if _, err := io.Copy(t, f); err != nil {
            http.Error(w, err.Error(),
            http.StatusInternalServerError)
            return
        }
        http.Redirect(w, r, "/view?id="+filename,
        http.StatusFound)
    }
}



func viewHandler(w http.ResponseWriter, r *http.Request) {
    imageId := r.FormValue("id")
    imagePath := UPLOAD_DIR + "/" + imageId
    if exists := isExists(imagePath);!exists {
        http.NotFound(w, r)
        return
    }
    w.Header().Set("Content-Type", "image")
    http.ServeFile(w, r, imagePath)
}
func isExists(path string) bool {
    _, err := os.Stat(path)
    if err == nil {
        return true
    }
    return os.IsExist(err)
}
func listHandler(w http.ResponseWriter, r *http.Request) {
    fileInfoArr, err := ioutil.ReadDir("./uploads")//"./uploads"
   /* if err != nil {
        
        http.Error(w, err.Error(),
        http.StatusInternalServerError)
        return
        
        
    }*/
    check(err)
    /*
    var listHtml string
    for _, fileInfo := range fileInfoArr {
       
        imgid := fileInfo.Name()
        listHtml += "<li><a href=\"/view?id="+imgid+"\">imgid</a></li>" 
    }
    io.WriteString(w, "<ol>"+listHtml+"</ol>")
    */
    locals := make(map[string]interface{})
    images := []string{}
    for _, fileInfo := range fileInfoArr {
        images = append(images, fileInfo.Name())
    }
    /*
    locals["images"] = images t, err := template.ParseFiles("list.html")
    if err != nil {
        http.Error(w, err.Error(),
        http.StatusInternalServerError)
        return
    }
    t.Execute(w, locals)*/
    locals["images"] = images
    if err = renderHtml(w, "list.html", locals); err != nil {
        /*
        http.Error(w, err.Error(),
        http.StatusInternalServerError)
        return
        */
        check(err)
    }
}
func renderHtml(w http.ResponseWriter, tmpl string, locals map[string]interface{})(err error){
    /*t, err = template.ParseFiles(tmpl + ".html")
    if err != nil {
        return
    }
    err = t.Execute(w, locals)
    */
    err = templates[tmpl].Execute(w, locals)
    return
}
func check(err error) {
    if err != nil {
        panic(err)
    }
}

func safeHandler(fn http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        defer func() {
            if e, ok := recover().(error); ok {
               // http.Error(w, err.Error(), http.StatusInternalServerError) //err 未定义
                // 或者输出自定义的50x错误页面
                // w.WriteHeader(http.StatusInternalServerError)
                // renderHtml(w, "error", e)
                // logging
                log.Println("WARN: panic in %v - %v", fn, e)
                log.Println(string(debug.Stack()))
            }
        }()
        fn(w, r)
    }
}
func staticDirHandler(mux *http.ServeMux, prefix string, staticDir string, flags int){
    mux.HandleFunc(prefix, func(w http.ResponseWriter, r *http.Request) {
        file := staticDir + r.URL.Path[len(prefix)-1:]
        if (flags & ListDir) == 0 {
            if exists := isExists(file); !exists {
                http.NotFound(w, r)
                return
            }
        }
        http.ServeFile(w, r, file)
    })
}
func main() {
   /*
    //不用safeHandler直接实现的情况下的代码
    http.HandleFunc("/", listHandler)
    http.HandleFunc("/view", viewHandler)
    http.HandleFunc("/upload", uploadHandler)
    
    //动态请求和静态资源分离的代码，运行失败
    mux := http.NewServeMux()
    staticDirHandler(mux, "/assets/", "./public", 0)*/
    http.HandleFunc("/", safeHandler(listHandler))
    http.HandleFunc("/view", safeHandler(viewHandler))
    http.HandleFunc("/upload", safeHandler(uploadHandler))
    
    err := http.ListenAndServe(":8090", nil)
    if err != nil {
     log.Fatal("ListenAndServe: ", err.Error())
    }
}
