package mongodb

import (
	"context"
	"reflect"

	"github.com/xhaoh94/gox/app"
	"github.com/xhaoh94/gox/engine/xlog"
	"github.com/xhaoh94/gox/util"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	client *mongo.Client
)

//ApplyDefaultUrl 连接默认数据库 mongodb://127.0.0.1:27017
func ApplyDefaultUrl() {
	url := "mongodb://"
	appCfg := app.GetAppCfg()
	if appCfg.MongoDb.User != "" && appCfg.MongoDb.Password != "" {
		url += appCfg.MongoDb.User + "@" + appCfg.MongoDb.Password + ":"
	}
	url += appCfg.MongoDb.Url
	ApplyURI(url)
}

//ApplyURI 连接数据库
func ApplyURI(url string) {
	// 设置客户端连接配置
	clientOptions := options.Client().ApplyURI(url)
	// 连接到MongoDB
	var err error
	client, err = mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		xlog.Fatal("MongoDB 连接失败[%v]", err)
	}
	// 检查连接
	err = client.Ping(context.TODO(), nil)
	if err != nil {
		xlog.Fatal("MongoDB 连接失败[%v]", err)
	}
	xlog.Info("成功连接 MongoDB! -> [%s]", url)
}

//GetClient 获取mongo客户端
func GetClient() *mongo.Client {
	return client
}

//GetDefaultDatabase 默认database
func GetDefaultDatabase() *mongo.Database {
	return client.Database(app.GetAppCfg().MongoDb.Database)
}

//CountDocuments 查询数量
func CountDocuments(cctn *mongo.Collection, filter interface{}) int64 {
	//filter模板 bson.M{"Email": ds.Email}
	count, err := cctn.CountDocuments(context.TODO(), filter)
	if err != nil {
		return -1
	}
	return count
}

//PushOne 添加一个
func PushOne(cctn *mongo.Collection, doc interface{}) string {
	rs, err := cctn.InsertOne(context.TODO(), doc)
	if err != nil {
		xlog.Error("MongoDB 添加单个错误[%v] 对象[%v]", err, doc)
		return ""
	}
	return (rs.InsertedID.(primitive.ObjectID)).Hex()
}

//PushMany 添加多个
func PushMany(cctn *mongo.Collection, docs []interface{}) []string {
	rs, err := cctn.InsertMany(context.TODO(), docs)
	if err != nil {
		xlog.Error("MongoDB 添加多个错误[%v] 对象[%v]", err, docs)
		return nil
	}
	result := make([]string, 0)
	for i := range rs.InsertedIDs {
		v := rs.InsertedIDs[i]
		result = append(result, (v.(primitive.ObjectID)).Hex())
	}
	return result
}

//GetOne 查询一个 filter[查询条件] out[返回查询数据]
func GetOne(cctn *mongo.Collection, filter interface{}, out interface{}) bool {
	t := reflect.TypeOf(out)
	if t.Kind() != reflect.Ptr {
		xlog.Error("MongoDB 查询对象需要指针类型[%v]", out)
		return false
	}
	err := cctn.FindOne(context.TODO(), filter).Decode(out)
	if err != nil {
		xlog.Error("MongoDB 查询单个错误[%v]", err)
		return false
	}
	return true
}

//GetMany 查询多个 filter[查询条件] out[返回查询的列表] findOptions[查询选项]
func GetMany(cctn *mongo.Collection, filter interface{}, out interface{}, findOptions *options.FindOptions) {
	t := reflect.TypeOf(out)
	if t.Kind() != reflect.Ptr {
		xlog.Error("MongoDB 查询对象需要指针类型[%v]", out)
		return
	}
	t1 := t.Elem()
	if t1.Kind() != reflect.Slice {
		xlog.Error("MongoDB 查询多个对象out参数必须是切片")
		return
	}
	childType := t1.Elem()
	if childType == nil {
		xlog.Error("MongoDB 查询多个对象切片子类型不可为空")
		return
	}

	cur, err := cctn.Find(context.TODO(), filter, findOptions)
	if err != nil {
		xlog.Error("MongoDB 查询多个错误[%v]", err)
		return
	}

	v := reflect.ValueOf(out).Elem()
	r := make([]reflect.Value, 0)

	// 查找多个文档返回一个光标
	// 遍历游标允许我们一次解码一个文档
	for cur.Next(context.TODO()) {
		// 创建一个值，将单个文档解码为该值
		elem := util.RTypeToInterface(childType)
		err := cur.Decode(elem)
		if err != nil {
			xlog.Error("MongoDB 查询多个错误[%v]", err)
			continue
		}
		r = append(r, reflect.ValueOf(elem))
	}
	if err := cur.Err(); err != nil {
		xlog.Error("MongoDB 查询多个错误[%v]", err)
	}
	// 完成后关闭游标
	cur.Close(context.TODO())
	v2 := reflect.Append(v, r...)
	v.Set(v2)
}

//DelOne 删除一个 filter[删除条件]
func DelOne(cctn *mongo.Collection, filter interface{}) int64 {
	// 删除名字是小黄的那个
	rs, err := cctn.DeleteOne(context.TODO(), filter)
	if err != nil {
		xlog.Error("MongoDB 删除单个错误[%v]", err)
		return 0
	}
	return rs.DeletedCount
}

//DelMany 删除多个  filter[删除条件] delOptions[删除选项]
func DelMany(cctn *mongo.Collection, filter interface{}, delOptions *options.DeleteOptions) int64 {
	rs, err := cctn.DeleteMany(context.TODO(), filter, delOptions)
	if err != nil {
		xlog.Error("MongoDB 删除多个错误[%v]", err)
		return 0
	}
	return rs.DeletedCount
}

//UpdateOne 更新单个 filter[查询条件] update[更新条件]
func UpdateOne(cctn *mongo.Collection, filter interface{}, update interface{}) (int64, int64, int64, interface{}) {
	rs, err := cctn.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		xlog.Error("MongoDB 更新单个错误[%v]", err)
		return 0, 0, 0, nil
	}
	return rs.MatchedCount, rs.ModifiedCount, rs.UpsertedCount, rs.UpsertedID
}

//UpdateMany 更新多个 filter[查询条件] update[更新条件]
func UpdateMany(cctn *mongo.Collection, filter interface{}, update interface{}) (int64, int64, int64, interface{}) {
	rs, err := cctn.UpdateMany(context.TODO(), filter, update)
	if err != nil {
		xlog.Error("MongoDB 更新多个错误[%v]", err)
		return 0, 0, 0, nil
	}
	return rs.MatchedCount, rs.ModifiedCount, rs.UpsertedCount, rs.UpsertedID
}
