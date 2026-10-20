package common

import (
	"errors"
	"image"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"
)

type ValidateArticleInput struct {
	Title       string
	Information string
	Content     string
	Tags        []string
}

type ValidateArticleOutput struct {
	Title       string
	Information string
	Content     string
	Tags        []string
}

var (
	TitleMaxLen       = 20
	InformationMaxLen = 200
	ContentMaxLen     = 50000
	CoverMaxLen       = 20
	TagMaxNumber      = 10
	TagMaxLen         = 10
	PageMinLen        = 1
	PageSizeMinLen    = 5
	PageSizeMaxLen    = 15
	QueryMaxLen       = 20
)

var (
	ErrTitleEmpty                = errors.New("标题为空！")
	ErrTitleOutRange             = errors.New("标题超出限制字数！")
	ErrInformationEmpty          = errors.New("简介为空！")
	ErrInformationOutRange       = errors.New("简介超出字数限制！")
	ErrContentEmpty              = errors.New("文章正文为空！")
	ErrContentOutRange           = errors.New("文章正文超出字数限制！")
	ErrCoverEmpty                = errors.New("封面属性为空！")
	ErrCoverOutRange             = errors.New("封面属性超出了字数限制！")
	ErrorCoverHaveSpecialContent = errors.New("封面属性含义特殊字符！")
	ErrTagsOutNumberRange        = errors.New("标签个数超出限制！")
	ErrTagEmpty                  = errors.New("标签为空！")
	ErrTagOutRange               = errors.New("标签内容超出限制！")
	ErrTagDuplicated             = errors.New("标签重复出现！")
	ErrPageOutRange              = errors.New("页数大小超出范围！")
	ErrArticleID                 = errors.New("文章ID格式有误！")
	ErrQueryEmpty                = errors.New("查询参数为空")
	ErrQueryOutRange             = errors.New("查询参数超出了范围")
)

// ValidatePage 校验page、size
func (c *CommonUtils) ValidatePage(page, pageSize int32) (int, int, error) {
	pg := int(page)
	pgs := int(pageSize)
	if pg >= PageMinLen && pgs >= PageSizeMinLen && pgs <= PageSizeMaxLen {
		return pg, pgs, nil
	}
	return 0, 0, ErrPageOutRange
}

// ValidateQuery 校验searchArticle.go中的参数query 返回一个模糊匹配字符串
func (c *CommonUtils) ValidateQuery(query string) (string, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return "", ErrQueryEmpty
	}
	if utf8.RuneCountInString(query) > QueryMaxLen {
		return "", ErrQueryOutRange
	}

	like := "%" + query + "%"
	return like, nil
}

// ValidateArticle 校验文章
func (c *CommonUtils) ValidateArticle(in *ValidateArticleInput) (*ValidateArticleOutput, error) {
	title, err := c.CheckTitle(in.Title)
	if err != nil {
		return nil, err
	}

	information, err := c.CheckInformation(in.Information)
	if err != nil {
		return nil, err
	}

	content, err := c.CheckContent(in.Content)
	if err != nil {
		return nil, err
	}

	//cover, err := c.CheckCover(in.Cover)
	//if err != nil {
	//	return nil, err
	//}

	tags, err := c.CheckTags(in.Tags)
	if err != nil {
		return nil, err
	}

	return &ValidateArticleOutput{
		Title:       title,
		Information: information,
		Content:     content,
		Tags:        tags,
	}, nil
}

// CheckTitle 校验标题
func (c *CommonUtils) CheckTitle(title string) (string, error) {
	title = strings.TrimSpace(title)
	if title == "" {
		return "", ErrTitleEmpty
	}
	if utf8.RuneCountInString(title) > TitleMaxLen {
		return "", ErrTitleOutRange
	}
	return title, nil
}

// CheckInformation 校验文章题记（主题介绍）
func (c *CommonUtils) CheckInformation(information string) (string, error) {
	information = strings.TrimSpace(information)
	if information == "" {
		return "", ErrInformationEmpty
	}
	if utf8.RuneCountInString(information) > InformationMaxLen {
		return "", ErrInformationOutRange
	}
	return information, nil
}

// CheckContent 校验文章正文
func (c *CommonUtils) CheckContent(content string) (string, error) {
	content = strings.TrimSpace(content)
	if content == "" {
		return "", ErrContentEmpty
	}
	if utf8.RuneCountInString(content) > ContentMaxLen {
		return "", ErrContentOutRange
	}
	return content, nil
}

// 只允许英文字母和数字，长度 1-128
var coverPattern = regexp.MustCompile(`^[a-zA-Z0-9]+$`)

// CheckCover 检查封面
func (c *CommonUtils) CheckCover(cover string) (string, error) {
	cover = strings.TrimSpace(cover)
	if cover == "" {
		return "", ErrCoverEmpty
	}
	if len(cover) > CoverMaxLen {
		return "", ErrCoverOutRange
	}
	if !coverPattern.MatchString(cover) {
		return "", ErrorCoverHaveSpecialContent
	}
	return cover, nil
}

// CheckTags 检查标签 返回去重、去空格后的字符串切片
func (c *CommonUtils) CheckTags(tags []string) ([]string, error) {
	if len(tags) <= 0 || len(tags) > TagMaxNumber {
		return nil, ErrTagsOutNumberRange
	}
	seen := make(map[string]struct{}, len(tags))
	result := make([]string, 0, len(tags))
	for _, tag := range tags {
		tag = strings.TrimSpace(tag)
		if tag == "" {
			return nil, ErrTagEmpty
		}
		if utf8.RuneCountInString(tag) > TagMaxLen {
			return nil, ErrTagOutRange
		}
		if _, dup := seen[tag]; dup {
			return nil, ErrTagDuplicated
		}
		seen[tag] = struct{}{}
		result = append(result, tag)
	}
	return result, nil
}

const (
	epoch          = int64(1735689600000)
	timestampShift = 22
)

// CheckArticleID 检查文章ID
// todo: 优化
func (c *CommonUtils) CheckArticleID(articleID string) (bool, error) {
	if len(articleID) <= 0 {
		return false, ErrArticleID
	}

	// 将string转为int64类型
	id, err := c.ToInt64(articleID)
	if err != nil { // 错误一般就是含有非数字元素 就当成格式错误了
		return false, ErrArticleID
	}

	timeStamp := id >> timestampShift
	now := time.Now().UnixMilli() - epoch // 毫秒级
	if timeStamp <= 0 || timeStamp > now+5*60*1000 {
		return false, ErrArticleID
	}
	return true, nil
}

var maxSize = 5 << 20
var allowedExt = map[string]struct{}{
	".jpg":  {},
	".jpeg": {},
	".png":  {},
	".webp": {},
}

// CheckImageFile 检查图片类型文件
func (c *CommonUtils) CheckImageFile(fileName, fileDir string) (bool, error) {
	if fileName == "" || fileDir == "" {
		return false, errors.New("文件不存在！")
	}

	ext := strings.ToLower(filepath.Ext(fileName))
	if _, ok := allowedExt[ext]; !ok {
		return false, errors.New("文件仅支持jpg/jpeg/png/webp！")
	}

	// 拼路径
	fp := filepath.Join("static/fileDir", fileName)

	absDir, err := filepath.Abs(fileDir)
	if err != nil {
		return false, errors.New("目录路径无效！")
	}

	absFile, err := filepath.Abs(fp)
	if err != nil {
		return false, errors.New("文件路径无效！")
	}

	if !strings.HasPrefix(absFile, absDir+string(os.PathSeparator)) {
		return false, errors.New("非法路径！")
	}

	// 检验文件是否存在
	st, err := os.Stat(absFile)
	if err != nil {
		if os.IsNotExist(err) {
			return false, errors.New("文件不存在！")
		}
		return false, errors.New("读取文件信息失败：" + err.Error())
	}
	if st.IsDir() {
		return false, errors.New("目标不是文件！")
	}

	// 大小校验
	if st.Size() <= 0 {
		return false, errors.New("文件为空！")
	}
	if st.Size() > int64(maxSize) {
		return false, errors.New("文件超过了5MB！")
	}

	if err := c.checkIsImage(absFile, ext); err != nil {
		return false, err
	}

	return true, nil
}

// checkIsImage 检验一个文件是否是图片
func (c *CommonUtils) checkIsImage(path, ext string) error {
	f, err := os.Open(path)
	if err != nil {
		return errors.New("打开文件失败：" + err.Error())
	}
	defer f.Close()

	// DecodeConfig 只读文件头，不解码整张图，性能友好
	_, format, err := image.DecodeConfig(f)
	if err != nil {
		return errors.New("不是有效的图片文件！")
	}

	// 扩展名与真实格式对齐（jpg和jpeg视为同一格式）
	want := strings.TrimPrefix(ext, ".")
	if want == "jpg" {
		want = "jpeg"
	}
	if want != format {
		return errors.New("文件扩展名与实际内容不符！")
	}
	return nil
}

func (c *CommonUtils) CheckArticleCover(cover, articleID string) (bool, error) {
	if cover == "" || articleID == "" {
		return false, errors.New("封面传入为空！")
	}
	var coverPrefix = "cover"
	if !strings.HasPrefix(cover, coverPrefix) {
		return false, errors.New("封面传入有误！")
	}

	// 得到cover后的雪花ID
	aid := strings.TrimPrefix(cover, coverPrefix)
	if aid == "" {
		return false, errors.New("封面传入有误！")
	}
	// 校验提取出的雪花ID与要修改的文章雪花ID是否相同
	if aid != articleID {
		return false, errors.New("封面传入有误！")
	}
	return true, nil
}
