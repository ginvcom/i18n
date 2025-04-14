package i18n

import (
	"context"
	"errors"
	"fmt"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/zeromicro/go-zero/core/logx"
	"golang.org/x/text/language"
	"google.golang.org/grpc/metadata"
	"gopkg.in/yaml.v3"
)

var bundle *i18n.Bundle

func RegisterLanguages(languages ...string) {
	if languages == nil {
		panic("at least one language pack is required")
	}

	bundle = i18n.NewBundle(language.English)
	ext := "yaml"
	bundle.RegisterUnmarshalFunc(ext, yaml.Unmarshal)
	for _, v := range languages {
		filePath := fmt.Sprintf("./etc/i18n/%s.%s", v, ext)
		bundle.MustLoadMessageFile(filePath)
	}

}

func Localize(lang string, msgId string, templateData ...interface{}) (msg string, err error) {
	if lang == "" {
		lang = "en"
	}

	localizer := i18n.NewLocalizer(bundle, lang)

	if templateData == nil {
		msg, err = localizer.Localize(&i18n.LocalizeConfig{
			MessageID: msgId,
		})
		if err != nil {
			return "", fmt.Errorf("localization failed: %w", err)
		}
	} else {
		msg, err = localizer.Localize(&i18n.LocalizeConfig{
			MessageID:    msgId,
			TemplateData: templateData[0],
		})

		if err != nil {
			return "", fmt.Errorf("localization failed with template data: %w", err)
		}
	}

	return
}

// A Lang represents a lang.
type Lang interface {
	Localize(msgId string, templateData ...interface{}) string
	Error(removedPrefixMsgId string, templateData ...interface{}) error
}

type lang struct {
	ctx context.Context
	// lang string
}

// WithContext sets ctx to lang, for keeping tracing information.
func WithContext(ctx context.Context) Lang {
	return &lang{
		ctx: ctx,
	}
}

// Localize retrieves the localized message for the given message ID and template data
// It first tries to get the language from the context metadata
// If no language is found, it defaults to English and returns an error message
// Returns the localized message string
func (l *lang) Localize(msgId string, templateData ...interface{}) string {
	md, ok := metadata.FromIncomingContext(l.ctx)
	if !ok {
		msg, err := Localize("en", "error.missingBasicParameter", map[string]string{"param": "lang"})
		if err != nil {
			logx.WithContext(l.ctx).Errorf("localization failed: %v", err)
			return err.Error()
		}

		return msg
	}

	// get lang
	lang := ""
	envArr := md.Get("lang")
	if len(envArr) >= 1 {
		lang = envArr[0]
	}

	msg, err := Localize(lang, msgId, templateData...)
	if err != nil {
		logx.WithContext(l.ctx).Errorf("localization failed: %v", err)
		return err.Error()
	}

	return msg
}

// Error creates a new error with a localized error message
// It prepends "error." prefix to the message ID and uses Localize to get the message
// Returns an error with the localized message
func (l *lang) Error(removedPrefixMsgId string, templateData ...interface{}) error {
	prefix := "error."
	msg := l.Localize(prefix+removedPrefixMsgId, templateData...)
	return errors.New(msg)
}
