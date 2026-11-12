// Package service policyreg：训练数据标注（需求 3.8-6）
// 口径：预标=实体词典/规则结果回标分词（token 完整包含于实体名→实体类型，否则 O）；
// 关系预标=机构×文件号同文档共现→「发文机构」；人工改标置 Source=人工（不可逆回规则）。
package service

import (
	"strings"

	"github.com/hequan2017/new-ops/server/plugin/plugin-tool/errx"
	"github.com/hequan2017/new-ops/server/plugin/policyreg/model"
)

// 错误码（policyreg 段 3000-3099）
const (
	ErrCodePrLabelParam    = 3006
	ErrCodePrLabelNotFound = 3007
)

// PreLabelTokens 实体预标（纯函数，测试锚点）：tokens 对齐 seq，命中实体名给类型
func PreLabelTokens(tokens []string, entities []model.PolicyEntity) []model.PolicyTokenLabel {
	labels := make([]model.PolicyTokenLabel, 0, len(tokens))
	for i, tk := range tokens {
		label := "O"
		for _, e := range entities {
			if e.Name != "" && strings.Contains(e.Name, tk) {
				label = e.EType
				break
			}
		}
		labels = append(labels, model.PolicyTokenLabel{
			Token: tk, Label: label, Source: "规则预标", Seq: i,
		})
	}
	return labels
}

// PreLabelRelations 关系预标（纯函数，测试锚点）：机构×文件号 共现
func PreLabelRelations(entities []model.PolicyEntity) []model.PolicyRelationLabel {
	out := make([]model.PolicyRelationLabel, 0)
	var orgs, docNos []string
	for _, e := range entities {
		switch e.EType {
		case "机构":
			orgs = append(orgs, e.Name)
		case "文件号":
			docNos = append(docNos, e.Name)
		}
	}
	for _, org := range orgs {
		for _, no := range docNos {
			out = append(out, model.PolicyRelationLabel{
				Head: org, Tail: no, Relation: "发文机构", Source: "规则预标",
			})
		}
	}
	return out
}

// BuildPolicyLabels 预标（按 docId：抽取实体→分词→回标，先清后插幂等）
func (s *PolicyService) BuildPolicyLabels(docId uint) (map[string]any, error) {
	doc, err := s.repo.FindDoc(s.dbh(), docId)
	if err != nil {
		return nil, errx.New(ErrCodePrLabelNotFound, "制度不存在")
	}
	if strings.TrimSpace(doc.RawText) == "" {
		return nil, errx.New(ErrCodePrLabelParam, "二进制编目文档无可标注文本")
	}
	entities := extractPolicyEntities(doc.RawText)
	tokens := PolicySegment(doc.RawText)
	tokenLabels := PreLabelTokens(tokens, entities)
	relations := PreLabelRelations(entities)
	if err := s.repo.DeleteTokenLabelsByDoc(s.dbh(), docId); err != nil {
		return nil, errx.Wrap(ErrCodePrStore, "清理分词标注失败", err)
	}
	if err := s.repo.DeleteRelationLabelsByDoc(s.dbh(), docId); err != nil {
		return nil, errx.Wrap(ErrCodePrStore, "清理关系标注失败", err)
	}
	for i := range tokenLabels {
		tokenLabels[i].DocId = docId
		if err := s.repo.CreateTokenLabel(s.dbh(), &tokenLabels[i]); err != nil {
			return nil, errx.Wrap(ErrCodePrStore, "创建分词标注失败", err)
		}
	}
	for i := range relations {
		relations[i].DocId = docId
		if err := s.repo.CreateRelationLabel(s.dbh(), &relations[i]); err != nil {
			return nil, errx.Wrap(ErrCodePrStore, "创建关系标注失败", err)
		}
	}
	return map[string]any{"docId": docId, "tokens": len(tokenLabels), "relations": len(relations)}, nil
}

// UpdatePolicyTokenLabel 人工改标
func (s *PolicyService) UpdatePolicyTokenLabel(id uint, label string) (*model.PolicyTokenLabel, error) {
	valid := map[string]bool{"O": true, "机构": true, "文件号": true, "日期": true, "专有名词": true}
	if !valid[label] {
		return nil, errx.New(ErrCodePrLabelParam, "标签须为 O/机构/文件号/日期/专有名词")
	}
	rec, err := s.repo.FindTokenLabel(s.dbh(), id)
	if err != nil {
		return nil, errx.New(ErrCodePrLabelNotFound, "标注不存在")
	}
	if err := s.repo.UpdateTokenLabelColumns(s.dbh(), rec.ID, map[string]any{
		"label": label, "source": "人工",
	}); err != nil {
		return nil, errx.Wrap(ErrCodePrStore, "更新标注失败", err)
	}
	rec.Label = label
	rec.Source = "人工"
	return rec, nil
}

// GetPolicyTokenLabels 分词标注列表
func (s *PolicyService) GetPolicyTokenLabels(docId uint) ([]model.PolicyTokenLabel, error) {
	list, err := s.repo.TokenLabelListByDoc(s.dbh(), docId)
	if err != nil {
		return nil, errx.Wrap(ErrCodePrStore, "查询分词标注失败", err)
	}
	return list, nil
}

// GetPolicyRelationLabels 关系标注列表
func (s *PolicyService) GetPolicyRelationLabels(docId uint) ([]model.PolicyRelationLabel, error) {
	list, err := s.repo.RelationLabelListByDoc(s.dbh(), docId)
	if err != nil {
		return nil, errx.Wrap(ErrCodePrStore, "查询关系标注失败", err)
	}
	return list, nil
}
