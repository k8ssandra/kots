package base

import (
	"github.com/pkg/errors"
	kotsv1beta1 "github.com/replicatedhq/kots/kotskinds/apis/kots/v1beta1"
	"github.com/replicatedhq/kots/pkg/crypto"
	"github.com/replicatedhq/kots/pkg/template"
	upstreamtypes "github.com/replicatedhq/kots/pkg/upstream/types"
)

func NewConfigContextTemplateBuilder(u *upstreamtypes.Upstream, renderOptions *RenderOptions) (*template.Builder, map[string]template.ItemValue, error) {
	kotsKinds, err := getKotsKinds(u, renderOptions.Log)
	if err != nil {
		return nil, nil, err
	}

	var templateContext map[string]template.ItemValue
	if kotsKinds.ConfigValues != nil {
		ctx := map[string]template.ItemValue{}
		for k, v := range kotsKinds.ConfigValues.Spec.Values {
			ctx[k] = template.ItemValue{
				Value:          v.Value,
				Default:        v.Default,
				Filename:       v.Filename,
				RepeatableItem: v.RepeatableItem,
			}
		}
		templateContext = ctx
	} else {
		templateContext = map[string]template.ItemValue{}
	}

	var cipher *crypto.AESCipher
	if u.EncryptionKey != "" {
		c, err := crypto.AESCipherFromString(u.EncryptionKey)
		if err != nil {
			return nil, nil, errors.Wrap(err, "failed to create cipher")
		}
		cipher = c
	}

	configGroups := []kotsv1beta1.ConfigGroup{}
	if kotsKinds.Config != nil {
		configGroups = kotsKinds.Config.Spec.Groups
	}

	localRegistry := template.LocalRegistry{
		Host:      renderOptions.LocalRegistryHost,
		Namespace: renderOptions.LocalRegistryNamespace,
		Username:  renderOptions.LocalRegistryUsername,
		Password:  renderOptions.LocalRegistryPassword,
		ReadOnly:  renderOptions.LocalRegistryIsReadOnly,
	}

	appInfo := template.ApplicationInfo{
		Slug: renderOptions.AppSlug,
	}

	versionInfo := template.VersionInfo{
		Sequence:     renderOptions.Sequence,
		Cursor:       u.UpdateCursor,
		ChannelName:  u.ChannelName,
		VersionLabel: u.VersionLabel,
		ReleaseNotes: u.ReleaseNotes,
		IsAirgap:     renderOptions.IsAirgap,
	}

	builderOptions := template.BuilderOptions{
		ConfigGroups:    configGroups,
		ExistingValues:  templateContext,
		LocalRegistry:   localRegistry,
		Cipher:          cipher,
		License:         kotsKinds.License,
		Application:     &kotsKinds.KotsApplication,
		VersionInfo:     &versionInfo,
		ApplicationInfo: &appInfo,
		IdentityConfig:  kotsKinds.IdentityConfig,
		Namespace:       renderOptions.Namespace,
	}
	builder, itemValues, err := template.NewBuilder(builderOptions)
	if err != nil {
		return &builder, itemValues, errors.Wrap(err, "failed to create config context")
	}

	return &builder, itemValues, nil
}
