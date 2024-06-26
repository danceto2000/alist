package fs

import (
	"context"
	"fmt"

	"github.com/alist-org/alist/v3/internal/driver"
	"github.com/alist-org/alist/v3/internal/errs"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/internal/op"
	"github.com/alist-org/alist/v3/pkg/tache"
	"github.com/pkg/errors"
)

type UploadTask struct {
	tache.Base
	Name             string `json:"name"`
	Status           string `json:"status"`
	storage          driver.Driver
	dstDirActualPath string
	file             model.FileStreamer
}

// func (t *UploadTask) OnFailed() {
// 	result := fmt.Sprintf("%s上传失败:%s", t.file.GetName(), t.GetErr())
// 	log.Debug(result)
// 	go op.Notify("文件上传结果", result)
// }

// func (t *UploadTask) OnSucceeded() {
// 	result := fmt.Sprintf("%s上传成功", t.file.GetName())
// 	log.Debug(result)
// 	go op.Notify("文件上传结果", "文件复制成功")
// }

func (t *UploadTask) GetName() string {
	return t.Name
	//return fmt.Sprintf("upload %s to [%s](%s)", t.file.GetName(), t.storage.GetStorage().MountPath, t.dstDirActualPath)
}

func (t *UploadTask) GetStatus() string {
	return t.Status
	//return "uploading"
}

func (t *UploadTask) Run() error {
	return op.Put(t.Ctx(), t.storage, t.dstDirActualPath, t.file, t.SetProgress, true)
}

var UploadTaskManager *tache.Manager[*UploadTask]

// putAsTask add as a put task and return immediately
func putAsTask(dstDirPath string, file model.FileStreamer) (tache.TaskWithInfo, error) {
	storage, dstDirActualPath, err := op.GetStorageAndActualPath(dstDirPath)
	if err != nil {
		return nil, errors.WithMessage(err, "failed get storage")
	}
	if storage.Config().NoUpload {
		return nil, errors.WithStack(errs.UploadNotSupported)
	}
	if file.NeedStore() {
		_, err := file.CacheFullInTempFile()
		if err != nil {
			return nil, errors.Wrapf(err, "failed to create temp file")
		}
		//file.SetReader(tempFile)
		//file.SetTmpFile(tempFile)
	}
	t := &UploadTask{
		Name:             fmt.Sprintf("upload %s to [%s](%s)", file.GetName(), storage.GetStorage().MountPath, dstDirActualPath),
		Status:           "uploading",
		storage:          storage,
		dstDirActualPath: dstDirActualPath,
		file:             file,
	}
	UploadTaskManager.Add(t)
	return t, nil
}

// putDirect put the file and return after finish
func putDirectly(ctx context.Context, dstDirPath string, file model.FileStreamer, lazyCache ...bool) error {
	storage, dstDirActualPath, err := op.GetStorageAndActualPath(dstDirPath)
	if err != nil {
		return errors.WithMessage(err, "failed get storage")
	}
	if storage.Config().NoUpload {
		return errors.WithStack(errs.UploadNotSupported)
	}
	return op.Put(ctx, storage, dstDirActualPath, file, nil, lazyCache...)
}
