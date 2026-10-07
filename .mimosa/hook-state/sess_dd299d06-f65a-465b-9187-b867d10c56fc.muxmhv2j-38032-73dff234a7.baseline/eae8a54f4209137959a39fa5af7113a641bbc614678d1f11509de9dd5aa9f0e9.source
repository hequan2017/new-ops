// Package service 白泽容器管理：镜像管理（M4→M8 C2 提前，场42）
// 镜像实时查询不落库；拉取为后台任务（内存状态表可查），完成即出现在列表。
package service

import (
	"context"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
)

// ImageView 镜像列表视图
type ImageView struct {
	ID        string   `json:"id"`
	Tags      []string `json:"tags"`
	SizeMB    int64    `json:"sizeMb"`
	CreatedAt int64    `json:"createdAt"`
}

// pullStatus 拉取任务状态（内存表：endpoint:ref → 状态）
type pullStatus struct {
	mu     sync.Mutex
	tasks  map[string]string // key → 拉取中/成功/失败: 原因
}

var imagePulls = &pullStatus{tasks: map[string]string{}}

func (p *pullStatus) set(key, status string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.tasks[key] = status
}

func (p *pullStatus) get(key string) string {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.tasks[key]
}

// ListImages 镜像列表（实时）
func (s *EndpointService) ListImages(endpointID uint) ([]ImageView, error) {
	ep, err := s.GetEndpoint(endpointID)
	if err != nil {
		return nil, err
	}
	cl, err := s.NewDockerClient(ep)
	if err != nil {
		return nil, err
	}
	defer cl.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	list, err := cl.ImageList(ctx, image.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("镜像列表拉取失败: %w", err)
	}
	out := make([]ImageView, 0, len(list))
	for _, im := range list {
		out = append(out, ImageView{
			ID:        shortID(strings.TrimPrefix(im.ID, "sha256:")),
			Tags:      im.RepoTags,
			SizeMB:    im.Size / 1024 / 1024,
			CreatedAt: im.Created,
		})
	}
	return out, nil
}

// RemoveImage 删除镜像（force）
func (s *EndpointService) RemoveImage(endpointID uint, imageRef string) error {
	if err := s.ensureEndpointAlive(endpointID); err != nil {
		return err
	}
	ep, err := s.GetEndpoint(endpointID)
	if err != nil {
		return err
	}
	cl, err := s.NewDockerClient(ep)
	if err != nil {
		return err
	}
	defer cl.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	_, err = cl.ImageRemove(ctx, imageRef, image.RemoveOptions{Force: true})
	return err
}

// PullImage 异步拉取镜像（内存状态表可查）
func (s *EndpointService) PullImage(endpointID uint, ref string) error {
	if strings.TrimSpace(ref) == "" {
		return newCtErr(ErrCodeEpAddrInvalid, "镜像引用不能为空")
	}
	key := fmt.Sprintf("%d:%s", endpointID, ref)
	if st := imagePulls.get(key); strings.HasPrefix(st, "拉取中") {
		return newCtErr(ErrCodeEpAddrInvalid, "该镜像正在拉取中")
	}
	ep, err := s.GetEndpoint(endpointID)
	if err != nil {
		return err
	}
	imagePulls.set(key, "拉取中")
	go func() {
		cl, err := s.NewDockerClient(ep)
		if err != nil {
			imagePulls.set(key, "失败: "+err.Error())
			return
		}
		defer cl.Close()
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
		defer cancel()
		reader, err := cl.ImagePull(ctx, ref, image.PullOptions{})
		if err != nil {
			imagePulls.set(key, "失败: "+err.Error())
			return
		}
		defer reader.Close()
		_, _ = io.Copy(io.Discard, reader) // 消费进度流至结束
		imagePulls.set(key, "成功 "+time.Now().Format("15:04:05"))
	}()
	return nil
}

// GetPullStatus 拉取状态查询
func (s *EndpointService) GetPullStatus(endpointID uint, ref string) string {
	return imagePulls.get(fmt.Sprintf("%d:%s", endpointID, ref))
}

// TagImage 为已有镜像打新标签
func (s *EndpointService) TagImage(endpointID uint, source, target string) error {
	if strings.TrimSpace(source) == "" || strings.TrimSpace(target) == "" {
		return newCtErr(ErrCodeEpAddrInvalid, "源与目标引用不能为空")
	}
	if err := s.ensureEndpointAlive(endpointID); err != nil {
		return err
	}
	ep, err := s.GetEndpoint(endpointID)
	if err != nil {
		return err
	}
	cl, err := s.NewDockerClient(ep)
	if err != nil {
		return err
	}
	defer cl.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	return cl.ImageTag(ctx, source, target)
}

// SaveImage 导出镜像为 tar（写入 writer；调用方控制 HTTP 下载流）
func (s *EndpointService) SaveImage(endpointID uint, refs []string, w io.Writer) error {
	if len(refs) == 0 {
		return newCtErr(ErrCodeEpAddrInvalid, "镜像引用不能为空")
	}
	ep, err := s.GetEndpoint(endpointID)
	if err != nil {
		return err
	}
	cl, err := s.NewDockerClient(ep)
	if err != nil {
		return err
	}
	defer cl.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	reader, err := cl.ImageSave(ctx, refs)
	if err != nil {
		return fmt.Errorf("导出失败: %w", err)
	}
	defer reader.Close()
	_, err = io.Copy(w, reader)
	return err
}

// LoadImage 导入镜像 tar（src 为上传文件流）
func (s *EndpointService) LoadImage(endpointID uint, src io.Reader) (string, error) {
	ep, err := s.GetEndpoint(endpointID)
	if err != nil {
		return "", err
	}
	cl, err := s.NewDockerClient(ep)
	if err != nil {
		return "", err
	}
	defer cl.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	resp, err := cl.ImageLoad(ctx, src, client.ImageLoadWithQuiet(false))
	if err != nil {
		return "", fmt.Errorf("导入失败: %w", err)
	}
	defer resp.Body.Close()
	out, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	return strings.TrimSpace(string(out)), nil
}
