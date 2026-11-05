package validate

import (
	"errors"
	"fmt"
	"image"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode/utf8"
)

type identityValidate interface {
	ValidateUserID(userID string) (string, error)
}
type ArticleValidate struct {
	identityValidate
}

func NewArticleValidate(identityValidate identityValidate) *ArticleValidate {
	return &ArticleValidate{
		identityValidate,
	}
}

const (
	TitleMaxLen       = 20
	InformationMaxLen = 200
	ContentMaxLen     = 50000
	TagMaxNumber      = 10
	TagMaxLen         = 10
	PageMinLen        = 1
	PageSizeMinLen    = 5
	PageSizeMaxLen    = 15
	QueryMaxLen       = 20
	MaxFileSize       = 5 << 20
)

var (
	ErrSnowFlakeIDInvalid      = errors.New("ID格式错误！")
	ErrArticleTitleEmpty       = errors.New("标题为空！")
	ErrTitleOutRange           = errors.New("标题超出限制字数！")
	ErrArticleInformationEmpty = errors.New("简介为空！")
	ErrInformationOutRange     = errors.New("简介超出字数限制！")
	ErrArticleContentEmpty     = errors.New("文章正文为空！")
	ErrContentOutRange         = errors.New("文章正文超出字数限制！")
	ErrTagsOutNumberRange      = errors.New("标签个数超出限制！")
	ErrTagEmpty                = errors.New("标签为空！")
	ErrTagOutRange             = errors.New("标签内容超出限制！")
	ErrTagDuplicated           = errors.New("标签重复出现！")
	ErrPageOutRange            = errors.New("页数大小超出范围！")
	ErrQueryEmpty              = errors.New("查询参数为空")
	ErrQueryOutRange           = errors.New("查询参数超出了范围")
)

type InputValidateArticle struct {
	Title       string
	Information string
	Content     string
	Tags        []string
}

type OutputValidateArticle struct {
	Title       string
	Information string
	Content     string
	Tags        []string
}

// ValidateArticle 校验文章
func (av *ArticleValidate) ValidateArticle(in InputValidateArticle) (OutputValidateArticle, error) {
	title, err := av.CheckTitle(in.Title)
	if err != nil {
		return OutputValidateArticle{}, err
	}

	information, err := av.CheckInformation(in.Information)
	if err != nil {
		return OutputValidateArticle{}, err
	}

	content, err := av.CheckContent(in.Content)
	if err != nil {
		return OutputValidateArticle{}, err
	}

	tags, err := av.CheckTags(in.Tags)
	if err != nil {
		return OutputValidateArticle{}, err
	}

	return OutputValidateArticle{
		Title:       title,
		Information: information,
		Content:     content,
		Tags:        tags,
	}, nil
}

type InputValidateUserArticleList struct {
	UserID   string
	Page     int32
	PageSize int32
}

type OuputValidateUserArticleList struct {
	UserID   string
	Page     int
	PageSize int
}

func (av *ArticleValidate) ValidateUserArticleList(in InputValidateUserArticleList) (OuputValidateUserArticleList, error) {
	page, pageSize, err := av.ValidatePage(in.Page, in.PageSize)
	if err != nil {
		return OuputValidateUserArticleList{}, err
	}

	userID, err := av.ValidateUserID(in.UserID)
	if err != nil {
		return OuputValidateUserArticleList{}, err
	}

	return OuputValidateUserArticleList{
		Page:     page,
		PageSize: pageSize,
		UserID:   userID,
	}, nil
}

const (
	epoch          = int64(1735689600000)
	timestampShift = 22
	workerIDShift  = 15
	maxWorkID      = 127
)

// ValidateSnowID 校验雪花生成器生产的ID 如文章ID、评论ID
func (av *ArticleValidate) ValidateSnowID(id int64) (int64, error) {
	if id <= 0 {
		return 0, ErrSnowFlakeIDInvalid
	}

	timestamp := id >> timestampShift
	if timestamp < 0 { // 时间戳不能早于epoch 即timestamp<0则表示该ID不是本项目生产
		return 0, ErrSnowFlakeIDInvalid
	}

	// 时间戳不能超过当前时间
	now := time.Now().UnixMilli()
	if epoch+timestamp > now+1000 {
		return 0, ErrSnowFlakeIDInvalid
	}

	// 提取wokerID并检查范围
	wokerID := (id >> workerIDShift) & maxWorkID
	if wokerID < 0 || wokerID > maxWorkID {
		return 0, ErrSnowFlakeIDInvalid
	}

	return id, nil
}

// ValidatePage 校验page、size
func (av *ArticleValidate) ValidatePage(page, pageSize int32) (int, int, error) {
	pg := int(page)
	pgs := int(pageSize)
	if pg >= PageMinLen && pgs >= PageSizeMinLen && pgs <= PageSizeMaxLen {
		return pg, pgs, nil
	}
	return 0, 0, ErrPageOutRange
}

// ValidateQuery 校验searchArticle.go中的参数query 返回一个模糊匹配字符串
func (av *ArticleValidate) ValidateQuery(query string) (string, error) {
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

// CheckTitle 校验标题
func (av *ArticleValidate) CheckTitle(title string) (string, error) {
	title = strings.TrimSpace(title)
	if title == "" {
		return "", ErrArticleTitleEmpty
	}
	if utf8.RuneCountInString(title) > TitleMaxLen {
		return "", ErrTitleOutRange
	}
	return title, nil
}

// CheckInformation 校验文章题记（主题介绍）
func (av *ArticleValidate) CheckInformation(information string) (string, error) {
	information = strings.TrimSpace(information)
	if information == "" {
		return "", ErrArticleInformationEmpty
	}
	if utf8.RuneCountInString(information) > InformationMaxLen {
		return "", ErrInformationOutRange
	}
	return information, nil
}

// CheckContent 校验文章正文
func (av *ArticleValidate) CheckContent(content string) (string, error) {
	content = strings.TrimSpace(content)
	if content == "" {
		return "", ErrArticleContentEmpty
	}
	if utf8.RuneCountInString(content) > ContentMaxLen {
		return "", ErrContentOutRange
	}
	return content, nil
}

// CheckTags 检查标签 返回去重、去空格后的字符串切片
func (av *ArticleValidate) CheckTags(tags []string) ([]string, error) {
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

var allowedExt = map[string]struct{}{
	".jpg":  {},
	".jpeg": {},
	".png":  {},
	".webp": {},
}

// CheckImageFile 检查图片类型数据 如文章封面、头像
func (av *ArticleValidate) CheckImageFile(fileName, fileDir string) (bool, error) {
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
	if st.Size() > int64(MaxFileSize) {
		return false, errors.New("文件超过了5MB！")
	}

	if err := av.checkIsImage(absFile, ext); err != nil {
		return false, err
	}

	return true, nil
}

// checkIsImage 检验一个文件是否是图片
func (av *ArticleValidate) checkIsImage(path, ext string) error {
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

func (av *ArticleValidate) CheckArticleCover(cover string, articleID int64) (bool, error) {
	if cover == "" || articleID <= 0 {
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
	//// 校验提取出的雪花ID与要修改的文章雪花ID是否相同
	//if aid != articleID {
	//	return false, errors.New("封面传入有误！")
	//}
	return true, nil
}

// SaveImageFile 保存文章封面
func (av *ArticleValidate) SaveImageFile(header *multipart.FileHeader, file multipart.File,
	saveDir string, articleID int64) (string, error) {
	// 文件校验
	if header == nil || file == nil {
		return "", errors.New("未上传有效文件！")
	}
	if header.Size <= 0 {
		return "", errors.New("未上传有效文件！")
	}
	if header.Size > int64(MaxFileSize) {
		return "", errors.New("上传文件超过了5MB！")
	}

	ext := strings.ToLower(filepath.Ext(header.Filename))
	if _, ok := allowedExt[ext]; !ok {
		return "", errors.New("上传文件仅支持jpg/jpeg/png/webp！")
	}

	// 将封面保存到static/articlecover中 保存名cover+articleID
	sd := filepath.Join("./static", saveDir)
	if err := os.MkdirAll(sd, 0755); err != nil {
		return "", errors.New("文件上传失败！")
	}

	fn := fmt.Sprintf("cover%dtime%d%s", articleID, time.Now().Unix(), ext)
	savePath := filepath.Join(sd, fn)

	// 每次写入前删除旧文件
	matches, _ := filepath.Glob(filepath.Join(sd, fmt.Sprintf("cover%dtime*.*", articleID)))
	for _, m := range matches {
		_ = os.Remove(m)
	}
	dst, err := os.Create(savePath)
	if err != nil {
		return "", fmt.Errorf("保存文件失败！%w", err)
	}

	// 写入失败时删除残余file
	if _, err := io.Copy(dst, file); err != nil {
		_ = dst.Close()
		_ = os.Remove(savePath)
		return "", fmt.Errorf("文件保存失败！%w", err)
	}
	_ = dst.Close()

	return fn, nil
}
