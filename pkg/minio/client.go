package minio

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	"github.com/pkg/errors"

	rawMinio "github.com/minio/minio-go/v6"
)

// ACLType bucket/object ACL
type ACLType string

const (
	// ACLPrivate definition : private read and write
	ACLPrivate ACLType = "private"

	// ACLPublicRead definition : public read and private write
	ACLPublicRead ACLType = "public-read"

	// ACLPublicReadWrite definition : public read and public write
	ACLPublicReadWrite ACLType = "public-read-write"
)

var (
	ErrEndpointInvalidPAth = errors.New("OSS对象存储Endpoint配置不合法: 不能带有Path路径")
	ErrEndpointInvalid     = errors.New("OSS对象存储Endpoint配置不合法: 未解析到http方法和host")
)

// ObjectStoreService对象存储服务接口
type ObjectStoreClient interface {
	// PreSignedPutObject 根据对象key生成上传文件的完整url链接，供其他不拥有access AK的应用方上传对象存储资源
	PreSignedPutObject(objectKey string, duration time.Duration) (signedUrl *url.URL, err error)

	PreSignedPostObject(objectKey string, duration time.Duration) (signedUrl *url.URL, formData map[string]string, err error)

	// PreSignedGetObject 根据对象key生成下载文件的完整url链接，供其他不拥有access AK的应用方下载对象存储资源
	PreSignedGetObject(objectKey string, duration time.Duration, opt ...Option) (signedUrl *url.URL, err error)

	// PutObjectFromFile 上传本地文件至对象存储服务
	FPutObject(ctx context.Context, localPath, objectKey, contentType string) error

	// 上传文件至对象存储服务
	PutObject(ctx context.Context, objectKey string, reader io.Reader, objectSize int64, contentType string) error

	// PutPrivateObjectFromFile 上传本地文件为私有对象存储
	FPutPrivateObject(ctx context.Context, localPath, objectKey, contentType string) error

	// FGetObject 下载对象文件到本地存储
	FGetObject(ctx context.Context, localPath, objectKey string) error

	// 下载文件
	GetObject(ctx context.Context, objectKey string) (io.Reader, error)

	// DeletePrefixObjectsExclude 删除prefixKey开头的所有的文件，只保留排除的excludeKey对应的对象存储文件
	DeletePrefixObjectsExclude(ctx context.Context, prefixKey, excludeKey string) error

	// ListObjects 获取指定目录下的对象列表
	ListObjects(prefix string, recursive bool, n int64) (objsInfo []rawMinio.ObjectInfo, err error)

	// ListPaginateObjects 获取指定目录下的对象分页列表
	ListPaginateObjects(prefix, keyWord string, page, pageSize, n int64, recursive bool) (result PaginateObjectsResp, err error)

	// DeleteObjects 删除指定key
	DeleteObjects(ctx context.Context, keys []string) error
}

type minioClient struct {
	cfg       *ossConfig
	rawClient *rawMinio.Client
}

func NewMinioClient(endpoint, accessID, accessSecret, bucketName, rootPath string) (ObjectStoreClient, error) {
	cfg, err := newOssConfig(endpoint, accessID, accessSecret, bucketName, rootPath)
	if err != nil {
		return nil, err
	}
	rawClient, err := rawMinio.New(cfg.Endpoint, cfg.AccessID, cfg.AccessSecret, cfg.Secure)
	if err != nil {
		return nil, err
	}
	return &minioClient{
		cfg:       cfg,
		rawClient: rawClient,
	}, nil
}

type ossConfig struct {
	Endpoint     string
	AccessID     string
	AccessSecret string
	RootPath     string
	BucketName   string
	Secure       bool
}

type PaginateObjectsResp struct {
	Items []rawMinio.ObjectInfo
	Total int64
}

func getSchemaAndHost(endpoint string) (schema, host string, err error) {
	parsedUrl, err1 := url.Parse(endpoint)
	if err1 != nil {
		err = errors.Wrapf(err, "解析OSS对象存储Endpoint错误.")
		return
	}

	//if parsedUrl.Path != "" && parsedUrl.Path != "/" {
	//	err = ErrEndpointInvalidPAth
	//	return
	//}
	if parsedUrl.Scheme == "" || parsedUrl.Host == "" {
		err = ErrEndpointInvalid
		return
	}
	return parsedUrl.Scheme, parsedUrl.Host, nil
}

func newOssConfig(endpoint, accessID, accessSecret, bucketName, rootPath string) (*ossConfig, error) {
	schema, host, err := getSchemaAndHost(endpoint)
	if err != nil {
		return nil, err
	}
	var secure bool
	if strings.EqualFold(schema, "https") {
		secure = true
	}
	return &ossConfig{
		Endpoint:     host,
		AccessID:     accessID,
		AccessSecret: accessSecret,
		BucketName:   bucketName,
		RootPath:     rootPath,
		Secure:       secure,
	}, nil
}

type Option func(val url.Values)

func WithContentDisposition(objectKey string) Option {
	return func(val url.Values) {
		_, fileName := filepath.Split(objectKey)
		val.Set("response-content-disposition", fmt.Sprintf("attachment; filename=%s", fileName))
	}
}

func WithRootPath(rootPath string, objectKey string) string {
	if rootPath != "" && !strings.HasPrefix(objectKey, rootPath) {
		objectKey = filepath.Join(rootPath, objectKey)
	}
	if strings.HasPrefix(objectKey, "/") {
		objectKey = strings.TrimPrefix(objectKey, "/")
	}
	return objectKey
}

func (c *minioClient) WithRootPath(objectKey string) string {
	return WithRootPath(c.cfg.RootPath, objectKey)
}

// PreSignedPutObject 根据对象key生成上传文件的完整url链接，供其他不拥有access AK的应用方上传对象存储资源
func (c *minioClient) PreSignedPutObject(objectKey string, duration time.Duration) (signedUrl *url.URL, err error) {
	return c.rawClient.PresignedPutObject(c.cfg.BucketName, c.WithRootPath(objectKey), duration)
}

// PreSignedPostObject 根据对象key生成上传文件的post url链接，供其他不拥有access AK的应用方上传对象存储资源
func (c *minioClient) PreSignedPostObject(objectKey string, duration time.Duration) (signedUrl *url.URL, formData map[string]string, err error) {
	postPolicy := rawMinio.NewPostPolicy()
	err = postPolicy.SetExpires(time.Now().Add(duration))
	if err != nil {
		return
	}
	err = postPolicy.SetKey(c.WithRootPath(objectKey))
	if err != nil {
		return
	}
	err = postPolicy.SetBucket(c.cfg.BucketName)
	if err != nil {
		return
	}
	return c.rawClient.PresignedPostPolicy(postPolicy)
}

// PreSignedGetObject 根据对象key生成下载文件的完整url链接，供其他不拥有access AK的应用方下载对象存储资源
func (c *minioClient) PreSignedGetObject(objectKey string, duration time.Duration, opt ...Option) (signedUrl *url.URL, err error) {
	reqParams := make(url.Values)
	for _, o := range opt {
		o(reqParams)
	}
	return c.rawClient.PresignedGetObject(c.cfg.BucketName, c.WithRootPath(objectKey), duration, reqParams)
}

// PutObjectFromFile 上传本地文件至对象存储服务
func (c *minioClient) FPutObject(ctx context.Context, localPath, objectKey, contentType string) error {
	if len(contentType) == 0 {
		contentType = "application/octet-stream"
	}
	putOptions := rawMinio.PutObjectOptions{ContentType: contentType}
	_, err := c.rawClient.FPutObjectWithContext(ctx, c.cfg.BucketName, c.WithRootPath(objectKey), localPath, putOptions)
	return err
}

func (c *minioClient) PutObject(ctx context.Context, objectKey string, reader io.Reader, objectSize int64, contentType string) error {
	if len(contentType) == 0 {
		contentType = "application/octet-stream"
	}
	putOptions := rawMinio.PutObjectOptions{ContentType: contentType}
	_, err := c.rawClient.PutObjectWithContext(ctx, c.cfg.BucketName, c.WithRootPath(objectKey), reader, objectSize, putOptions)
	return err
}

// PutPrivateObjectFromFile 上传本地文件为私有对象存储
func (c *minioClient) FPutPrivateObject(ctx context.Context, localPath, objectKey, contentType string) error {
	userMetadata := map[string]string{
		"x-amz-acl": string(ACLPrivate),
	}
	if len(contentType) == 0 {
		contentType = "application/octet-stream"
	}
	putOptions := rawMinio.PutObjectOptions{ContentType: contentType, UserMetadata: userMetadata}
	_, err := c.rawClient.FPutObjectWithContext(ctx, c.cfg.BucketName, c.WithRootPath(objectKey), localPath, putOptions)
	return err
}

// FGetObject 下载对象文件到本地存储
func (c *minioClient) FGetObject(ctx context.Context, localPath, objectKey string) error {
	getOptions := rawMinio.GetObjectOptions{}
	return c.rawClient.FGetObjectWithContext(ctx, c.cfg.BucketName, c.WithRootPath(objectKey), localPath, getOptions)
}

func (c *minioClient) GetObject(ctx context.Context, objectKey string) (io.Reader, error) {
	getOptions := rawMinio.GetObjectOptions{}
	return c.rawClient.GetObjectWithContext(ctx, c.cfg.BucketName, c.WithRootPath(objectKey), getOptions)
}

// DeleteObjectList 删除prefixKey开头的所有的文件，只保留排除的excludeKey对应的对象存储文件
func (c *minioClient) DeletePrefixObjectsExclude(ctx context.Context, prefixKey, excludeKey string) error {
	deleteKeysCh := make(chan string)
	done := make(chan error)

	go func() {
		var err error
		defer func() {
			close(deleteKeysCh)
			done <- err
		}()

		for object := range c.rawClient.ListObjects(c.cfg.BucketName, c.WithRootPath(prefixKey), true, nil) {
			if object.Err != nil {
				err = object.Err
				return
			}
			if excludeKey != "" && strings.Contains(object.Key, excludeKey) {
				continue
			}
			deleteKeysCh <- object.Key
		}
	}()

	for rErr := range c.rawClient.RemoveObjectsWithContext(ctx, c.cfg.BucketName, deleteKeysCh) {
		if rErr.Err != nil {
			return rErr.Err
		}
	}
	return <-done
}

// ListObjects 获取指定目录下的对象列表
func (c minioClient) ListObjects(prefix string, recursive bool, n int64) (objsInfo []rawMinio.ObjectInfo, err error) {

	// Create a done channel to control 'ListObjects' go routine.
	doneCh := make(chan struct{}, 1)

	// Free the channel upon return.
	defer close(doneCh)

	i := int64(1)
	for object := range c.rawClient.ListObjectsV2WithMetadata(c.cfg.BucketName, c.WithRootPath(prefix), recursive, doneCh) {
		if object.Err != nil {
			return
		}
		i++
		// Verify if we have printed N objects.
		if i == n && n != 0 {
			// Indicate ListObjects go-routine to exit and stop
			// feeding the objectInfo channel.
			doneCh <- struct{}{}
		}
		objsInfo = append(objsInfo, object)
	}
	return
}

// ListPaginateObjects 获取指定目录下的对象分页列表
func (c minioClient) ListPaginateObjects(prefix, keyWord string, page, pageSize, n int64, recursive bool) (result PaginateObjectsResp, err error) {
	objects, err := c.ListObjects(c.WithRootPath(prefix), recursive, n)
	if err != nil {
		return
	}

	var filteredObjects []rawMinio.ObjectInfo
	for _, object := range objects {
		realName := strings.Replace(object.Key, c.WithRootPath(prefix), "", -1)
		if keyWord == "" {
			filteredObjects = append(filteredObjects, object)
			continue
		} else if strings.Contains(realName, keyWord) {
			filteredObjects = append(filteredObjects, object)
		}
	}
	total := int64(len(filteredObjects))
	if page == 0 {
		page = 1
	}
	skip := (page - 1) * pageSize
	limit := func() int64 {
		if skip+pageSize > total {
			return total
		} else {
			return skip + pageSize
		}

	}
	start := func() int64 {
		if skip > total {
			return total
		} else {
			return skip
		}
	}
	result.Items = filteredObjects[start():limit()]
	result.Total = total
	return
}

// DeleteObjectList 删除指定的key
func (c *minioClient) DeleteObjects(ctx context.Context, keys []string) error {
	deleteKeysCh := make(chan string)
	done := make(chan error)

	go func() {
		var err error
		defer func() {
			close(deleteKeysCh)
			done <- err
		}()
		for _, key := range keys {
			deleteKeysCh <- c.WithRootPath(key)
		}
	}()

	for rErr := range c.rawClient.RemoveObjectsWithContext(ctx, c.cfg.BucketName, deleteKeysCh) {
		if rErr.Err != nil {
			return rErr.Err
		}
	}
	return <-done
}
