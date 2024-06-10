// package router
package cpython

import "C"
import (
	"fmt"
	"github.com/ccfos/nightingale/v6/models"
	"github.com/gin-gonic/gin"
	"github.com/toolkits/pkg/ginx"
	"net/http"
	"unsafe"
)

type Message struct {
	Name string `json:"name"`
	Body string `json:"body"`
}

// unc forecasting_main() {
// 	router := gin.Default()

// 	// 定义路由
// 	router.GET("/messages", getMessages)
// 	router.GET("/messages/:name", getMessage)
// 	router.POST("/messages", postMessage)
// 	router.DELETE("/messages/:name", deleteMessage)

// 	// 运行服务器
// 	router.Run(":8080")
//

// #cgo CFLAGS: -I"Q:/Sill-/anaconda/envs/poetry/include"
// #cgo LDFLAGS: -L"Q:/Sill-/anaconda/envs/poetry/libs" -lpython311
// #include <Python.h>

//import "C"

func init() {
	C.init_python()
}

func runPythonCode(code string) int {
	ccode := C.CString(code)
	defer C.free(unsafe.Pointer(ccode))
	return int(C.run_python_code(ccode))
}

func main() {
	code := `print("Hello from Python!")`
	if result := runPythonCode(code); result != 0 {
		fmt.Println("Error executing Python code")
	}
}

/*
#cgo pkg-config: python-3.10
#define Py_LIMITED_API
#include <Python.h>

extern void greetFunc(char *name, char *result);
void greetFunc(char *name, char *result) {
    // 初始化Python
    Py_Initialize();

    // 导入Python模块
    PyObject *pName = PyUnicode_FromString("example");
    PyObject *pModule = PyImport_Import(pName);
    Py_DECREF(pName);

    if (pModule != NULL) {
        // 获取模块中的函数
        PyObject *pFunc = PyObject_GetAttrString(pModule, "greet");

        if (pFunc && PyCallable_Check(pFunc)) {
            // 构造参数
            PyObject *pArgs = PyTuple_Pack(1, PyUnicode_FromString(name));
            PyObject *pValue = PyObject_CallObject(pFunc, pArgs);
            Py_DECREF(pArgs);

            if (pValue != NULL) {
                // 转换返回值为字符串
                const char *ret = PyUnicode_AsUTF8(pValue);
                strcpy(result, ret);
                Py_DECREF(pValue);
            } else {
                PyErr_Print();
            }
        } else {
            PyErr_Print();
        }

        Py_XDECREF(pFunc);
        Py_DECREF(pModule);
        Py_Finalize();
    }
}
*/

func main() {
	name := "World"
	var result [50]byte

	// 调用cgo函数
	C.greetFunc(C.CString(name), (*C.char)(unsafe.Pointer(&result)))

	// 打印结果
	fmt.Println(C.GoString((*C.char)(unsafe.Pointer(&result))))
}

// 获取所有消息
func getTainModelMessages(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "所有消息"})
}

// 根据名字获取消息
func getPredictResultMessage(c *gin.Context) {
	name := c.Param("name")
	c.JSON(http.StatusOK, gin.H{"message": "消息: " + name})
}

// 发送消息
func postMessage(c *gin.Context) {
	var message Message
	if c.BindJSON(&message) == nil {
		c.JSON(http.StatusCreated, gin.H{"message": message})
	}
}

// 删除消息
func deleteMessage(c *gin.Context) {
	name := c.Param("name")
	c.JSON(http.StatusOK, gin.H{"message": "删除消息: " + name})
}

// no param
func (rt *Router) metricViewGets(c *gin.Context) {
	lst, err := models.MetricViewGets(rt.Ctx, c.MustGet("userid"))
	ginx.NewRender(c).Data(lst, err)
}

// body: name, configs, cate
func (rt *Router) metricViewAdd(c *gin.Context) {
	var f models.MetricView
	ginx.BindJSON(c, &f)

	me := c.MustGet("user").(*models.User)
	if !me.IsAdmin() {
		// 管理员可以选择当前这个视图是公开呢，还是私有，普通用户的话就只能是私有的
		f.Cate = 1
	}

	f.Id = 0
	f.CreateBy = me.Id

	ginx.Dangerous(f.Add(rt.Ctx))

	ginx.NewRender(c).Data(f, nil)
}

// body: ids
func (rt *Router) metricViewDel(c *gin.Context) {
	var f idsForm
	ginx.BindJSON(c, &f)
	f.Verify()

	me := c.MustGet("user").(*models.User)
	if me.IsAdmin() {
		ginx.NewRender(c).Message(models.MetricViewDel(rt.Ctx, f.Ids))
	} else {
		ginx.NewRender(c).Message(models.MetricViewDel(rt.Ctx, f.Ids, me.Id))
	}
}

// body: id, name, configs, cate
func (rt *Router) metricViewPut(c *gin.Context) {
	var f models.MetricView
	ginx.BindJSON(c, &f)

	view, err := models.MetricViewGet(rt.Ctx, "id = ?", f.Id)
	ginx.Dangerous(err)

	if view == nil {
		ginx.NewRender(c).Message("no such item(id: %d)", f.Id)
		return
	}

	me := c.MustGet("user").(*models.User)
	if !me.IsAdmin() {
		f.Cate = 1

		// 如果是普通用户，只能修改自己的
		if view.CreateBy != me.Id {
			ginx.NewRender(c, http.StatusForbidden).Message("forbidden")
			return
		}
	}

	ginx.NewRender(c).Message(view.Update(rt.Ctx, f.Name, f.Configs, f.Cate, me.Id))
}
