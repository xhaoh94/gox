package mongodb

import (
	"context"
	"reflect"

	"github.com/xhaoh94/gox/engine/helper/commonhelper"
	"github.com/xhaoh94/gox/engine/logger"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MgoDb struct {
	client *mongo.Client
}

// ApplyDefaultUrl 连接默认数据库 mongodb://127.0.0.1:27017
// func ApplyDefaultUrl() {
// 	url := "mongodb://"
// 	appCfg := app.GetAppCfg()
// 	if appCfg.MongoDb.User != "" && appCfg.MongoDb.Password != "" {
// 		url += appCfg.MongoDb.User + "@" + appCfg.MongoDb.Password + ":"
// 	}
// 	url += appCfg.MongoDb.Url
// 	ApplyURI(url)
// }

// ApplyURI 连接数据库
func (db *MgoDb) ApplyURI(url string) {
	// 设置客户端连接配置
	clientOptions := options.Client().ApplyURI(url)
	// 连接到MongoDB
	var err error
	db.client, err = mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		logger.Fatal().Err(err).Msg("MongoDB 连接失败")
	}
	// 检查连接
	err = db.client.Ping(context.TODO(), nil)
	if err != nil {
		logger.Fatal().Err(err).Msg("MongoDB 连接失败")
	}
	logger.Info().Str("Url", url).Msg("MongoDB 连接成功")
}

// GetClient 获取mongo客户端
func (db *MgoDb) GetClient() *mongo.Client {
	return db.client
}

// GetDefaultDatabase 默认database
// func (db *MgoDb) GetDefaultDatabase() *mongo.Database {
// 	return db.client.Database(app.GetAppCfg().MongoDb.Database)
// }

// CountDocuments 查询数量
func (db *MgoDb) CountDocuments(cctn *mongo.Collection, filter interface{}) int64 {
	//filter模板 bson.M{"Email": ds.Email}
	count, err := cctn.CountDocuments(context.TODO(), filter)
	if err != nil {
		return -1
	}
	return count
}

// PushOne 添加一个
func (db *MgoDb) PushOne(cctn *mongo.Collection, doc interface{}) string {
	rs, err := cctn.InsertOne(context.TODO(), doc)
	if err != nil {
		logger.Error().Err(err).Interface("对象", doc).Msg("MongoDB 添加单个错误")
		return ""
	}
	return (rs.InsertedID.(primitive.ObjectID)).Hex()
}

// PushMany 添加多个
func (db *MgoDb) PushMany(cctn *mongo.Collection, docs []interface{}) []string {
	rs, err := cctn.InsertMany(context.TODO(), docs)
	if err != nil {
		logger.Error().Err(err).Interface("对象", docs).Msg("MongoDB 添加多个错误")
		return nil
	}
	result := make([]string, 0)
	for i := range rs.InsertedIDs {
		v := rs.InsertedIDs[i]
		result = append(result, (v.(primitive.ObjectID)).Hex())
	}
	return result
}

// GetOne 查询一个 filter[查询条件] out[返回查询数据]
func (db *MgoDb) GetOne(cctn *mongo.Collection, filter interface{}, out interface{}) bool {
	t := reflect.TypeOf(out)
	if t.Kind() != reflect.Ptr {
		logger.Error().Interface("对象", out).Msg("MongoDB 查询对象需要指针类型")
		return false
	}
	err := cctn.FindOne(context.TODO(), filter).Decode(out)
	if err != nil {
		logger.Error().Err(err).Msg("MongoDB 查询单个错误")
		return false
	}
	return true
}

// GetMany 查询多个 filter[查询条件] out[返回查询的列表] findOptions[查询选项]
func (db *MgoDb) GetMany(cctn *mongo.Collection, filter interface{}, out interface{}, findOptions *options.FindOptions) {
	t := reflect.TypeOf(out)
	if t.Kind() != reflect.Ptr {
		logger.Error().Interface("对象", out).Msg("MongoDB 查询对象需要指针类型")
		return
	}
	t1 := t.Elem()
	if t1.Kind() != reflect.Slice {
		logger.Error().Interface("对象", out).Msg("MongoDB 查询多个对象out参数必须是切片")
		return
	}
	childType := t1.Elem()
	if childType == nil {
		logger.Error().Interface("对象", out).Msg("MongoDB 查询多个对象切片子类型不可为空")
		return
	}

	cur, err := cctn.Find(context.TODO(), filter, findOptions)
	if err != nil {
		logger.Error().Err(err).Msg("MongoDB 查询多个错误")
		return
	}

	v := reflect.ValueOf(out).Elem()
	r := make([]reflect.Value, 0)

	// 查找多个文档返回一个光标
	// 遍历游标允许我们一次解码一个文档
	for cur.Next(context.TODO()) {
		// 创建一个值，将单个文档解码为该值
		elem := commonhelper.RTypeToInterface(childType)
		err := cur.Decode(elem)
		if err != nil {
			logger.Error().Err(err).Msg("MongoDB 查询多个错误")
			continue
		}
		r = append(r, reflect.ValueOf(elem))
	}
	if err := cur.Err(); err != nil {
		logger.Error().Err(err).Msg("MongoDB 查询多个错误")
	}
	// 完成后关闭游标
	cur.Close(context.TODO())
	v2 := reflect.Append(v, r...)
	v.Set(v2)
}

// DelOne 删除一个 filter[删除条件]
func (db *MgoDb) DelOne(cctn *mongo.Collection, filter interface{}) int64 {
	// 删除名字是小黄的那个
	rs, err := cctn.DeleteOne(context.TODO(), filter)
	if err != nil {
		logger.Error().Err(err).Msg("MongoDB 删除单个错误")
		return 0
	}
	return rs.DeletedCount
}

// DelMany 删除多个  filter[删除条件] delOptions[删除选项]
func (db *MgoDb) DelMany(cctn *mongo.Collection, filter interface{}, delOptions *options.DeleteOptions) int64 {
	rs, err := cctn.DeleteMany(context.TODO(), filter, delOptions)
	if err != nil {
		logger.Error().Err(err).Msg("MongoDB 删除多个错误")
		return 0
	}
	return rs.DeletedCount
}

// UpdateOne 更新单个 filter[查询条件] update[更新条件]
func (db *MgoDb) UpdateOne(cctn *mongo.Collection, filter interface{}, update interface{}) (int64, int64, int64, interface{}) {
	rs, err := cctn.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		logger.Error().Err(err).Msg("MongoDB 更新单个错误")
		return 0, 0, 0, nil
	}
	return rs.MatchedCount, rs.ModifiedCount, rs.UpsertedCount, rs.UpsertedID
}

// UpdateMany 更新多个 filter[查询条件] update[更新条件]
func (db *MgoDb) UpdateMany(cctn *mongo.Collection, filter interface{}, update interface{}) (int64, int64, int64, interface{}) {
	rs, err := cctn.UpdateMany(context.TODO(), filter, update)
	if err != nil {
		logger.Error().Err(err).Msg("MongoDB 更新多个错误")
		return 0, 0, 0, nil
	}
	return rs.MatchedCount, rs.ModifiedCount, rs.UpsertedCount, rs.UpsertedID
}
