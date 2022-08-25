package config

import (
	"errors"
	"strings"
	"testing"

	"github.com/prebid/prebid-server/errortypes"
	"github.com/prebid/prebid-server/openrtb_ext"
	"github.com/stretchr/testify/assert"
)

const testInfoFilesPath = "./test/bidder-info"
const testInvalidInfoFilesPath = "./test/bidder-info-invalid"

func TestLoadBidderInfoFromDisk(t *testing.T) {
	bidder := "someBidder"
	trueValue := true

	adapterConfigs := make(map[string]Adapter)
	adapterConfigs[strings.ToLower(bidder)] = Adapter{}

	infos, err := LoadBidderInfoFromDisk(testInfoFilesPath)
	if err != nil {
		t.Fatal(err)
	}

	expected := BidderInfos{
		bidder: {
			Disabled: false,
			Maintainer: &MaintainerInfo{
				Email: "some-email@domain.com",
			},
			GVLVendorID: 42,
			Capabilities: &CapabilitiesInfo{
				App: &PlatformInfo{
					MediaTypes: []openrtb_ext.BidType{openrtb_ext.BidTypeBanner, openrtb_ext.BidTypeNative},
				},
				Site: &PlatformInfo{
					MediaTypes: []openrtb_ext.BidType{openrtb_ext.BidTypeBanner, openrtb_ext.BidTypeVideo, openrtb_ext.BidTypeNative},
				},
			},
			Syncer: &Syncer{
				Key: "foo",
				IFrame: &SyncerEndpoint{
					URL:         "https://foo.com/sync?mode=iframe&r={{.RedirectURL}}",
					RedirectURL: "{{.ExternalURL}}/setuid/iframe",
					ExternalURL: "https://iframe.host",
					UserMacro:   "%UID",
				},
				Redirect: &SyncerEndpoint{
					URL:         "https://foo.com/sync?mode=redirect&r={{.RedirectURL}}",
					RedirectURL: "{{.ExternalURL}}/setuid/redirect",
					ExternalURL: "https://redirect.host",
					UserMacro:   "#UID",
				},
				SupportCORS: &trueValue,
			},
		},
	}
	assert.Equal(t, expected, infos)
}

func TestLoadBidderInfoInvalid(t *testing.T) {
	expectedError := "error parsing yaml for bidder someBidder-invalid.yaml: yaml: unmarshal errors:\n  line 3: cannot unmarshal !!str `42` into uint16"
	_, err := loadBidderInfo(testInvalidInfoFilesPath)
	assert.EqualError(t, err, expectedError, "incorrect error message returned while loading invalid bidder config")
}

func TestToGVLVendorIDMap(t *testing.T) {
	givenBidderInfos := BidderInfos{
		"bidderA": BidderInfo{Disabled: false, GVLVendorID: 0},
		"bidderB": BidderInfo{Disabled: false, GVLVendorID: 100},
		"bidderC": BidderInfo{Disabled: true, GVLVendorID: 0},
		"bidderD": BidderInfo{Disabled: true, GVLVendorID: 200},
	}

	expectedGVLVendorIDMap := map[openrtb_ext.BidderName]uint16{
		"bidderB": 100,
	}

	result := givenBidderInfos.ToGVLVendorIDMap()
	assert.Equal(t, expectedGVLVendorIDMap, result)
}

const bidderInfoRelativePath = "../static/bidder-info"

// TestBidderInfoFiles ensures each bidder has a valid static/bidder-info/bidder.yaml file. Validation is performed directly
// against the file system with separate yaml unmarshalling from the LoadBidderInfoFromDisk func.
func TestBidderInfoFiles(t *testing.T) {
	_, errs := ProcessBidderInfos(bidderInfoRelativePath, nil)
	if len(errs) > 0 {
		errorMsg := errortypes.NewAggregateError("bidder infos", errs)
		assert.Fail(t, errorMsg.Message, "Errors in bidder info files")
	}
}

func TestBidderInfoFilesWithCorrectConfigs(t *testing.T) {
	bidderInfos, errs := ProcessBidderInfos(bidderInfoRelativePath, BidderInfos{"appnexus": BidderInfo{
		Endpoint: "http://override.com",
	}})
	if len(errs) > 0 {
		errorMsg := errortypes.NewAggregateError("bidder infos", errs)
		assert.Fail(t, errorMsg.Message, "Errors in bidder info files")
	}
	assert.Equal(t, "http://override.com", bidderInfos["appnexus"].Endpoint, "Incorrect endpoint override")
}

func TestBidderInfoFilesWithIncorrectConfigs(t *testing.T) {
	_, errs := ProcessBidderInfos(bidderInfoRelativePath, BidderInfos{"appnexus": BidderInfo{
		Endpoint: "incorrect",
	}})
	assert.Len(t, errs, 1, "Incorrect error number returned from process bidder infos with invalid enpoint")
	assert.Equal(t, "The endpoint: incorrect for appnexus is not a valid URL", errs[0].Error(), "Incorrect error returned from process bidder infos with invalid enpoint")
}

func TestBidderInfoFilesWithIncorrectSyncerConfigs(t *testing.T) {
	_, errs := ProcessBidderInfos(bidderInfoRelativePath, BidderInfos{"adgeneration": BidderInfo{
		UserSyncURL: "usersyncurl",
	}})
	assert.Len(t, errs, 1, "Incorrect error number returned from process bidder infos with invalid enpoint")
	assert.Equal(t, "adapters.adgeneration.usersync_url cannot be applied, bidder does not define a user sync", errs[0].Error(), "Incorrect error returned from process bidder infos with invalid enpoint")
}

func TestBidderInfoValidationPositive(t *testing.T) {
	bidderInfos := BidderInfos{
		"bidderA": BidderInfo{
			Endpoint:   "http://bidderA.com/openrtb2",
			PlatformID: "A",
			Maintainer: &MaintainerInfo{
				Email: "maintainer@bidderA.com",
			},
			GVLVendorID: 1,
			Capabilities: &CapabilitiesInfo{
				App: &PlatformInfo{
					MediaTypes: []openrtb_ext.BidType{
						openrtb_ext.BidTypeVideo,
						openrtb_ext.BidTypeNative,
						openrtb_ext.BidTypeBanner,
					},
				},
			},
			Syncer: &Syncer{
				Key: "bidderAkey",
				Redirect: &SyncerEndpoint{
					URL:       "http://bidderA.com/usersync",
					UserMacro: "UID",
				},
			},
		},
		"bidderB": BidderInfo{
			Endpoint:   "http://bidderB.com/openrtb2",
			PlatformID: "B",
			Maintainer: &MaintainerInfo{
				Email: "maintainer@bidderA.com",
			},
			GVLVendorID: 2,
			Capabilities: &CapabilitiesInfo{
				Site: &PlatformInfo{
					MediaTypes: []openrtb_ext.BidType{
						openrtb_ext.BidTypeVideo,
						openrtb_ext.BidTypeNative,
						openrtb_ext.BidTypeBanner,
					},
				},
			},
			Syncer: &Syncer{
				Key: "bidderBkey",
				Redirect: &SyncerEndpoint{
					URL:       "http://bidderB.com/usersync",
					UserMacro: "UID",
				},
			},
		},
	}
	errs := validateBidderInfos(bidderInfos)
	assert.Len(t, errs, 0, "All bidder infos should be correct")
}

func TestBidderInfoValidationNegative(t *testing.T) {
	testCases := []struct {
		description  string
		bidderInfos  BidderInfos
		expectErrors []error
	}{
		{
			"One bidder incorrect url",
			BidderInfos{
				"bidderA": BidderInfo{
					Endpoint: "incorrect",
					Maintainer: &MaintainerInfo{
						Email: "maintainer@bidderA.com",
					},
					Capabilities: &CapabilitiesInfo{
						App: &PlatformInfo{
							MediaTypes: []openrtb_ext.BidType{
								openrtb_ext.BidTypeVideo,
							},
						},
					},
				},
			},
			[]error{
				errors.New("The endpoint: incorrect for bidderA is not a valid URL"),
			},
		},
		{
			"One bidder empty url",
			BidderInfos{
				"bidderA": BidderInfo{
					Endpoint: "",
					Maintainer: &MaintainerInfo{
						Email: "maintainer@bidderA.com",
					},
					Capabilities: &CapabilitiesInfo{
						App: &PlatformInfo{
							MediaTypes: []openrtb_ext.BidType{
								openrtb_ext.BidTypeVideo,
							},
						},
					},
				},
			},
			[]error{
				errors.New("There's no default endpoint available for bidderA. Calls to this bidder/exchange will fail. Please set adapters.bidderA.endpoint in your app config"),
			},
		},
		{
			"One bidder incorrect url template",
			BidderInfos{
				"bidderA": BidderInfo{
					Endpoint: "http://bidderA.com/openrtb2/getuid?{{.incorrect}}",
					Maintainer: &MaintainerInfo{
						Email: "maintainer@bidderA.com",
					},
					Capabilities: &CapabilitiesInfo{
						App: &PlatformInfo{
							MediaTypes: []openrtb_ext.BidType{
								openrtb_ext.BidTypeVideo,
							},
						},
					},
				},
			},
			[]error{
				errors.New("Unable to resolve endpoint: http://bidderA.com/openrtb2/getuid?{{.incorrect}} for adapter: bidderA. template: endpointTemplate:1:37: executing \"endpointTemplate\" at <.incorrect>: can't evaluate field incorrect in type macros.EndpointTemplateParams"),
			},
		},
		{
			"One bidder incorrect url template parameters",
			BidderInfos{
				"bidderA": BidderInfo{
					Endpoint: "http://bidderA.com/openrtb2/getuid?r=[{{.]RedirectURL}}",
					Maintainer: &MaintainerInfo{
						Email: "maintainer@bidderA.com",
					},
					Capabilities: &CapabilitiesInfo{
						App: &PlatformInfo{
							MediaTypes: []openrtb_ext.BidType{
								openrtb_ext.BidTypeVideo,
							},
						},
					},
				},
			},
			[]error{
				errors.New("Invalid endpoint template: http://bidderA.com/openrtb2/getuid?r=[{{.]RedirectURL}} for adapter: bidderA. template: endpointTemplate:1: bad character U+005D ']'"),
			},
		},
		{
			"One bidder no maintainer",
			BidderInfos{
				"bidderA": BidderInfo{
					Endpoint: "http://bidderA.com/openrtb2",
					Capabilities: &CapabilitiesInfo{
						App: &PlatformInfo{
							MediaTypes: []openrtb_ext.BidType{
								openrtb_ext.BidTypeVideo,
							},
						},
					},
				},
			},
			[]error{
				errors.New("missing required field: maintainer.email for adapter: bidderA"),
			},
		},
		{
			"One bidder missing maintainer email",
			BidderInfos{
				"bidderA": BidderInfo{
					Endpoint: "http://bidderA.com/openrtb2",
					Maintainer: &MaintainerInfo{
						Email: "",
					},
					Capabilities: &CapabilitiesInfo{
						App: &PlatformInfo{
							MediaTypes: []openrtb_ext.BidType{
								openrtb_ext.BidTypeVideo,
							},
						},
					},
				},
			},
			[]error{
				errors.New("missing required field: maintainer.email for adapter: bidderA"),
			},
		},
		{
			"One bidder missing capabilities",
			BidderInfos{
				"bidderA": BidderInfo{
					Endpoint: "http://bidderA.com/openrtb2",
					Maintainer: &MaintainerInfo{
						Email: "maintainer@bidderA.com",
					},
				},
			},
			[]error{
				errors.New("missing required field: capabilities for adapter: bidderA"),
			},
		},
		{
			"One bidder missing capabilities site and app",
			BidderInfos{
				"bidderA": BidderInfo{
					Endpoint: "http://bidderA.com/openrtb2",
					Maintainer: &MaintainerInfo{
						Email: "maintainer@bidderA.com",
					},
					Capabilities: &CapabilitiesInfo{},
				},
			},
			[]error{
				errors.New("at least one of capabilities.site or capabilities.app must exist for adapter: bidderA"),
			},
		},
		{
			"One bidder incorrect capabilities for app",
			BidderInfos{
				"bidderA": BidderInfo{
					Endpoint: "http://bidderA.com/openrtb2",
					Maintainer: &MaintainerInfo{
						Email: "maintainer@bidderA.com",
					},
					Capabilities: &CapabilitiesInfo{
						App: &PlatformInfo{
							MediaTypes: []openrtb_ext.BidType{
								"incorrect",
							},
						},
					},
				},
			},
			[]error{
				errors.New("capabilities.app failed validation: unrecognized media type at index 0: incorrect for adapter: bidderA"),
			},
		},
		{
			"One bidder empty capabilities for app",
			BidderInfos{
				"bidderA": BidderInfo{
					Endpoint: "http://bidderA.com/openrtb2",
					Maintainer: &MaintainerInfo{
						Email: "maintainer@bidderA.com",
					},
					Capabilities: &CapabilitiesInfo{
						App: &PlatformInfo{
							MediaTypes: []openrtb_ext.BidType{},
						},
					},
				},
			},
			[]error{
				errors.New("capabilities.app failed validation: mediaTypes should be an array with at least one string element for adapter: bidderA"),
			},
		},
		{
			"One bidder nil capabilities",
			BidderInfos{
				"bidderA": BidderInfo{
					Endpoint: "http://bidderA.com/openrtb2",
					Maintainer: &MaintainerInfo{
						Email: "maintainer@bidderA.com",
					},
					Capabilities: nil,
				},
			},
			[]error{
				errors.New("missing required field: capabilities for adapter: bidderA"),
			},
		},
		{
			"One bidder empty capabilities for site",
			BidderInfos{
				"bidderA": BidderInfo{
					Endpoint: "http://bidderA.com/openrtb2",
					Maintainer: &MaintainerInfo{
						Email: "maintainer@bidderA.com",
					},
					Capabilities: &CapabilitiesInfo{
						Site: &PlatformInfo{
							MediaTypes: []openrtb_ext.BidType{},
						},
					},
				},
			},
			[]error{
				errors.New("capabilities.site failed validation: mediaTypes should be an array with at least one string element, for adapter: bidderA"),
			},
		},
		{
			"One bidder invalid syncer",
			BidderInfos{
				"bidderA": BidderInfo{
					Endpoint: "http://bidderA.com/openrtb2",
					Maintainer: &MaintainerInfo{
						Email: "maintainer@bidderA.com",
					},
					Capabilities: &CapabilitiesInfo{
						Site: &PlatformInfo{
							MediaTypes: []openrtb_ext.BidType{
								openrtb_ext.BidTypeVideo,
							},
						},
					},
					Syncer: &Syncer{
						Supports: []string{"incorrect"},
					},
				},
			},
			[]error{
				errors.New("syncer could not be created, invalid supported endpoint: incorrect"),
			},
		},
		{
			"Two bidders, one with incorrect url",
			BidderInfos{
				"bidderA": BidderInfo{
					Endpoint: "incorrect",
					Maintainer: &MaintainerInfo{
						Email: "maintainer@bidderA.com",
					},
					Capabilities: &CapabilitiesInfo{
						App: &PlatformInfo{
							MediaTypes: []openrtb_ext.BidType{
								openrtb_ext.BidTypeVideo,
							},
						},
					},
				},
				"bidderB": BidderInfo{
					Endpoint: "http://bidderB.com/openrtb2",
					Maintainer: &MaintainerInfo{
						Email: "maintainer@bidderB.com",
					},
					Capabilities: &CapabilitiesInfo{
						App: &PlatformInfo{
							MediaTypes: []openrtb_ext.BidType{
								openrtb_ext.BidTypeVideo,
							},
						},
					},
				},
			},
			[]error{
				errors.New("The endpoint: incorrect for bidderA is not a valid URL"),
			},
		},
		{
			"Two bidders, both with incorrect url",
			BidderInfos{
				"bidderA": BidderInfo{
					Endpoint: "incorrect",
					Maintainer: &MaintainerInfo{
						Email: "maintainer@bidderA.com",
					},
					Capabilities: &CapabilitiesInfo{
						App: &PlatformInfo{
							MediaTypes: []openrtb_ext.BidType{
								openrtb_ext.BidTypeVideo,
							},
						},
					},
				},
				"bidderB": BidderInfo{
					Endpoint: "incorrect",
					Maintainer: &MaintainerInfo{
						Email: "maintainer@bidderB.com",
					},
					Capabilities: &CapabilitiesInfo{
						App: &PlatformInfo{
							MediaTypes: []openrtb_ext.BidType{
								openrtb_ext.BidTypeVideo,
							},
						},
					},
				},
			},
			[]error{
				errors.New("The endpoint: incorrect for bidderA is not a valid URL"),
				errors.New("The endpoint: incorrect for bidderB is not a valid URL"),
			},
		},
	}

	for _, test := range testCases {
		errs := validateBidderInfos(test.bidderInfos)
		assert.ElementsMatch(t, errs, test.expectErrors, "incorrect errors returned for test: %s", test.description)
	}
}

func TestSyncerOverride(t *testing.T) {
	var (
		trueValue  = true
		falseValue = false
	)

	testCases := []struct {
		description   string
		givenOriginal *Syncer
		givenOverride *Syncer
		expected      *Syncer
	}{
		{
			description:   "Nil",
			givenOriginal: nil,
			givenOverride: nil,
			expected:      nil,
		},
		{
			description:   "Original Nil",
			givenOriginal: nil,
			givenOverride: &Syncer{Key: "anyKey"},
			expected:      &Syncer{Key: "anyKey"},
		},
		{
			description:   "Original Empty",
			givenOriginal: &Syncer{},
			givenOverride: &Syncer{Key: "anyKey"},
			expected:      &Syncer{Key: "anyKey"},
		},
		{
			description:   "Override Nil",
			givenOriginal: &Syncer{Key: "anyKey"},
			givenOverride: nil,
			expected:      &Syncer{Key: "anyKey"},
		},
		{
			description:   "Override Empty",
			givenOriginal: &Syncer{Key: "anyKey"},
			givenOverride: &Syncer{},
			expected:      &Syncer{Key: "anyKey"},
		},
		{
			description:   "Override Key",
			givenOriginal: &Syncer{Key: "original"},
			givenOverride: &Syncer{Key: "override"},
			expected:      &Syncer{Key: "override"},
		},
		{
			description:   "Override IFrame",
			givenOriginal: &Syncer{IFrame: &SyncerEndpoint{URL: "original"}},
			givenOverride: &Syncer{IFrame: &SyncerEndpoint{URL: "override"}},
			expected:      &Syncer{IFrame: &SyncerEndpoint{URL: "override"}},
		},
		{
			description:   "Override Redirect",
			givenOriginal: &Syncer{Redirect: &SyncerEndpoint{URL: "original"}},
			givenOverride: &Syncer{Redirect: &SyncerEndpoint{URL: "override"}},
			expected:      &Syncer{Redirect: &SyncerEndpoint{URL: "override"}},
		},
		{
			description:   "Override ExternalURL",
			givenOriginal: &Syncer{ExternalURL: "original"},
			givenOverride: &Syncer{ExternalURL: "override"},
			expected:      &Syncer{ExternalURL: "override"},
		},
		{
			description:   "Override SupportCORS",
			givenOriginal: &Syncer{SupportCORS: &trueValue},
			givenOverride: &Syncer{SupportCORS: &falseValue},
			expected:      &Syncer{SupportCORS: &falseValue},
		},
		{
			description:   "Override Partial - Other Fields Untouched",
			givenOriginal: &Syncer{Key: "originalKey", ExternalURL: "originalExternalURL"},
			givenOverride: &Syncer{ExternalURL: "overrideExternalURL"},
			expected:      &Syncer{Key: "originalKey", ExternalURL: "overrideExternalURL"},
		},
	}

	for _, test := range testCases {
		result := test.givenOverride.Override(test.givenOriginal)
		assert.Equal(t, test.expected, result, test.description)
	}
}

func TestSyncerEndpointOverride(t *testing.T) {
	testCases := []struct {
		description   string
		givenOriginal *SyncerEndpoint
		givenOverride *SyncerEndpoint
		expected      *SyncerEndpoint
	}{
		{
			description:   "Nil",
			givenOriginal: nil,
			givenOverride: nil,
			expected:      nil,
		},
		{
			description:   "Original Nil",
			givenOriginal: nil,
			givenOverride: &SyncerEndpoint{URL: "anyURL"},
			expected:      &SyncerEndpoint{URL: "anyURL"},
		},
		{
			description:   "Original Empty",
			givenOriginal: &SyncerEndpoint{},
			givenOverride: &SyncerEndpoint{URL: "anyURL"},
			expected:      &SyncerEndpoint{URL: "anyURL"},
		},
		{
			description:   "Override Nil",
			givenOriginal: &SyncerEndpoint{URL: "anyURL"},
			givenOverride: nil,
			expected:      &SyncerEndpoint{URL: "anyURL"},
		},
		{
			description:   "Override Empty",
			givenOriginal: &SyncerEndpoint{URL: "anyURL"},
			givenOverride: &SyncerEndpoint{},
			expected:      &SyncerEndpoint{URL: "anyURL"},
		},
		{
			description:   "Override URL",
			givenOriginal: &SyncerEndpoint{URL: "original"},
			givenOverride: &SyncerEndpoint{URL: "override"},
			expected:      &SyncerEndpoint{URL: "override"},
		},
		{
			description:   "Override RedirectURL",
			givenOriginal: &SyncerEndpoint{RedirectURL: "original"},
			givenOverride: &SyncerEndpoint{RedirectURL: "override"},
			expected:      &SyncerEndpoint{RedirectURL: "override"},
		},
		{
			description:   "Override ExternalURL",
			givenOriginal: &SyncerEndpoint{ExternalURL: "original"},
			givenOverride: &SyncerEndpoint{ExternalURL: "override"},
			expected:      &SyncerEndpoint{ExternalURL: "override"},
		},
		{
			description:   "Override UserMacro",
			givenOriginal: &SyncerEndpoint{UserMacro: "original"},
			givenOverride: &SyncerEndpoint{UserMacro: "override"},
			expected:      &SyncerEndpoint{UserMacro: "override"},
		},
		{
			description:   "Override",
			givenOriginal: &SyncerEndpoint{URL: "originalURL", RedirectURL: "originalRedirectURL", ExternalURL: "originalExternalURL", UserMacro: "originalUserMacro"},
			givenOverride: &SyncerEndpoint{URL: "overideURL", RedirectURL: "overideRedirectURL", ExternalURL: "overideExternalURL", UserMacro: "overideUserMacro"},
			expected:      &SyncerEndpoint{URL: "overideURL", RedirectURL: "overideRedirectURL", ExternalURL: "overideExternalURL", UserMacro: "overideUserMacro"},
		},
	}

	for _, test := range testCases {
		result := test.givenOverride.Override(test.givenOriginal)
		assert.Equal(t, test.expected, result, test.description)
	}
}

func TestApplyBidderInfoConfigSyncerOverrides(t *testing.T) {
	var testCases = []struct {
		description            string
		givenFsBidderInfos     BidderInfos
		givenConfigBidderInfos BidderInfos
		expectedError          string
		expectedBidderInfos    BidderInfos
	}{
		{
			description:            "Syncer Override",
			givenFsBidderInfos:     BidderInfos{"a": {Syncer: &Syncer{Key: "original"}}},
			givenConfigBidderInfos: BidderInfos{"a": {Syncer: &Syncer{Key: "override"}}},
			expectedBidderInfos:    BidderInfos{"a": {Syncer: &Syncer{Key: "override"}}},
		},
		{
			description:            "UserSyncURL Override IFrame",
			givenFsBidderInfos:     BidderInfos{"a": {Syncer: &Syncer{IFrame: &SyncerEndpoint{URL: "original"}}}},
			givenConfigBidderInfos: BidderInfos{"a": {UserSyncURL: "override"}},
			expectedBidderInfos:    BidderInfos{"a": {UserSyncURL: "override", Syncer: &Syncer{IFrame: &SyncerEndpoint{URL: "override"}}}},
		},
		{
			description:            "UserSyncURL Supports IFrame",
			givenFsBidderInfos:     BidderInfos{"a": {Syncer: &Syncer{Supports: []string{"iframe"}}}},
			givenConfigBidderInfos: BidderInfos{"a": {UserSyncURL: "override"}},
			expectedBidderInfos:    BidderInfos{"a": {UserSyncURL: "override", Syncer: &Syncer{Supports: []string{"iframe"}, IFrame: &SyncerEndpoint{URL: "override"}}}},
		},
		{
			description:            "UserSyncURL Override Redirect",
			givenFsBidderInfos:     BidderInfos{"a": {Syncer: &Syncer{Supports: []string{"redirect"}}}},
			givenConfigBidderInfos: BidderInfos{"a": {UserSyncURL: "override"}},
			expectedBidderInfos:    BidderInfos{"a": {UserSyncURL: "override", Syncer: &Syncer{Supports: []string{"redirect"}, Redirect: &SyncerEndpoint{URL: "override"}}}},
		},
		{
			description:            "UserSyncURL Supports Redirect",
			givenFsBidderInfos:     BidderInfos{"a": {Syncer: &Syncer{Redirect: &SyncerEndpoint{URL: "original"}}}},
			givenConfigBidderInfos: BidderInfos{"a": {UserSyncURL: "override"}},
			expectedBidderInfos:    BidderInfos{"a": {UserSyncURL: "override", Syncer: &Syncer{Redirect: &SyncerEndpoint{URL: "override"}}}},
		},
		{
			description:            "UserSyncURL Override Syncer Not Defined",
			givenFsBidderInfos:     BidderInfos{"a": {}},
			givenConfigBidderInfos: BidderInfos{"a": {UserSyncURL: "override"}},
			expectedError:          "adapters.a.usersync_url cannot be applied, bidder does not define a user sync",
		},
		{
			description:            "UserSyncURL Override Syncer Endpoints Not Defined",
			givenFsBidderInfos:     BidderInfos{"a": {Syncer: &Syncer{}}},
			givenConfigBidderInfos: BidderInfos{"a": {UserSyncURL: "override"}},
			expectedError:          "adapters.a.usersync_url cannot be applied, bidder does not define user sync endpoints and does not define supported endpoints",
		},
		{
			description:            "UserSyncURL Override Ambiguous",
			givenFsBidderInfos:     BidderInfos{"a": {Syncer: &Syncer{IFrame: &SyncerEndpoint{URL: "originalIFrame"}, Redirect: &SyncerEndpoint{URL: "originalRedirect"}}}},
			givenConfigBidderInfos: BidderInfos{"a": {UserSyncURL: "override"}},
			expectedError:          "adapters.a.usersync_url cannot be applied, bidder defines multiple user sync endpoints or supports multiple endpoints",
		},
		{
			description:            "UserSyncURL Supports Ambiguous",
			givenFsBidderInfos:     BidderInfos{"a": {Syncer: &Syncer{Supports: []string{"iframe", "redirect"}}}},
			givenConfigBidderInfos: BidderInfos{"a": {UserSyncURL: "override"}},
			expectedError:          "adapters.a.usersync_url cannot be applied, bidder defines multiple user sync endpoints or supports multiple endpoints",
		},
	}

	for _, test := range testCases {
		bidderInfos, resultErr := applyBidderInfoConfigOverrides(test.givenConfigBidderInfos, test.givenFsBidderInfos)
		if test.expectedError == "" {
			assert.NoError(t, resultErr, test.description+":err")
			assert.Equal(t, test.expectedBidderInfos, bidderInfos, test.description+":result")
		} else {
			assert.EqualError(t, resultErr, test.expectedError, test.description+":err")
		}
	}
}

func TestApplyBidderInfoConfigOverrides(t *testing.T) {
	var testCases = []struct {
		description            string
		givenFsBidderInfos     BidderInfos
		givenConfigBidderInfos BidderInfos
		expectedError          string
		expectedBidderInfos    BidderInfos
	}{
		{
			description:            "Don't override endpoint",
			givenFsBidderInfos:     BidderInfos{"a": {Endpoint: "original"}},
			givenConfigBidderInfos: BidderInfos{"a": {Syncer: &Syncer{Key: "override"}}},
			expectedBidderInfos:    BidderInfos{"a": {Endpoint: "original", Syncer: &Syncer{Key: "override"}}},
		},
		{
			description:            "Override endpoint",
			givenFsBidderInfos:     BidderInfos{"a": {Endpoint: "original"}},
			givenConfigBidderInfos: BidderInfos{"a": {Endpoint: "override", Syncer: &Syncer{Key: "override"}}},
			expectedBidderInfos:    BidderInfos{"a": {Endpoint: "override", Syncer: &Syncer{Key: "override"}}},
		},
		{
			description:            "Don't override ExtraAdapterInfo",
			givenFsBidderInfos:     BidderInfos{"a": {ExtraAdapterInfo: "original"}},
			givenConfigBidderInfos: BidderInfos{"a": {Syncer: &Syncer{Key: "override"}}},
			expectedBidderInfos:    BidderInfos{"a": {ExtraAdapterInfo: "original", Syncer: &Syncer{Key: "override"}}},
		},
		{
			description:            "Override ExtraAdapterInfo",
			givenFsBidderInfos:     BidderInfos{"a": {ExtraAdapterInfo: "original"}},
			givenConfigBidderInfos: BidderInfos{"a": {ExtraAdapterInfo: "override", Syncer: &Syncer{Key: "override"}}},
			expectedBidderInfos:    BidderInfos{"a": {ExtraAdapterInfo: "override", Syncer: &Syncer{Key: "override"}}},
		},
		{
			description:            "Don't override Maintainer",
			givenFsBidderInfos:     BidderInfos{"a": {Maintainer: &MaintainerInfo{Email: "original"}}},
			givenConfigBidderInfos: BidderInfos{"a": {Syncer: &Syncer{Key: "override"}}},
			expectedBidderInfos:    BidderInfos{"a": {Maintainer: &MaintainerInfo{Email: "original"}, Syncer: &Syncer{Key: "override"}}},
		},
		{
			description:            "Override maintainer",
			givenFsBidderInfos:     BidderInfos{"a": {Maintainer: &MaintainerInfo{Email: "original"}}},
			givenConfigBidderInfos: BidderInfos{"a": {Maintainer: &MaintainerInfo{Email: "override"}, Syncer: &Syncer{Key: "override"}}},
			expectedBidderInfos:    BidderInfos{"a": {Maintainer: &MaintainerInfo{Email: "override"}, Syncer: &Syncer{Key: "override"}}},
		},
		{
			description: "Don't override Capabilities",
			givenFsBidderInfos: BidderInfos{"a": {
				Capabilities: &CapabilitiesInfo{App: &PlatformInfo{MediaTypes: []openrtb_ext.BidType{openrtb_ext.BidTypeVideo}}},
			}},
			givenConfigBidderInfos: BidderInfos{"a": {Syncer: &Syncer{Key: "override"}}},
			expectedBidderInfos: BidderInfos{"a": {
				Syncer:       &Syncer{Key: "override"},
				Capabilities: &CapabilitiesInfo{App: &PlatformInfo{MediaTypes: []openrtb_ext.BidType{openrtb_ext.BidTypeVideo}}},
			}},
		},
		{
			description: "Override Capabilities",
			givenFsBidderInfos: BidderInfos{"a": {
				Capabilities: &CapabilitiesInfo{App: &PlatformInfo{MediaTypes: []openrtb_ext.BidType{openrtb_ext.BidTypeVideo}}},
			}},
			givenConfigBidderInfos: BidderInfos{"a": {
				Syncer:       &Syncer{Key: "override"},
				Capabilities: &CapabilitiesInfo{App: &PlatformInfo{MediaTypes: []openrtb_ext.BidType{openrtb_ext.BidTypeBanner}}},
			}},
			expectedBidderInfos: BidderInfos{"a": {
				Syncer:       &Syncer{Key: "override"},
				Capabilities: &CapabilitiesInfo{App: &PlatformInfo{MediaTypes: []openrtb_ext.BidType{openrtb_ext.BidTypeBanner}}},
			}},
		},
		{
			description:            "Don't override Debug",
			givenFsBidderInfos:     BidderInfos{"a": {Debug: &DebugInfo{Allow: true}}},
			givenConfigBidderInfos: BidderInfos{"a": {Syncer: &Syncer{Key: "override"}}},
			expectedBidderInfos:    BidderInfos{"a": {Debug: &DebugInfo{Allow: true}, Syncer: &Syncer{Key: "override"}}},
		},
		{
			description:            "Override Debug",
			givenFsBidderInfos:     BidderInfos{"a": {Debug: &DebugInfo{Allow: true}}},
			givenConfigBidderInfos: BidderInfos{"a": {Debug: &DebugInfo{Allow: false}, Syncer: &Syncer{Key: "override"}}},
			expectedBidderInfos:    BidderInfos{"a": {Debug: &DebugInfo{Allow: false}, Syncer: &Syncer{Key: "override"}}},
		},
		{
			description:            "Don't override GVLVendorID",
			givenFsBidderInfos:     BidderInfos{"a": {GVLVendorID: 5}},
			givenConfigBidderInfos: BidderInfos{"a": {Syncer: &Syncer{Key: "override"}}},
			expectedBidderInfos:    BidderInfos{"a": {GVLVendorID: 5, Syncer: &Syncer{Key: "override"}}},
		},
		{
			description:            "Override GVLVendorID",
			givenFsBidderInfos:     BidderInfos{"a": {}},
			givenConfigBidderInfos: BidderInfos{"a": {GVLVendorID: 5, Syncer: &Syncer{Key: "override"}}},
			expectedBidderInfos:    BidderInfos{"a": {GVLVendorID: 5, Syncer: &Syncer{Key: "override"}}},
		},
		{
			description: "Don't override XAPI",
			givenFsBidderInfos: BidderInfos{"a": {
				XAPI: AdapterXAPI{Username: "username1", Password: "password2", Tracker: "tracker3"},
			}},
			givenConfigBidderInfos: BidderInfos{"a": {Syncer: &Syncer{Key: "override"}}},
			expectedBidderInfos: BidderInfos{"a": {
				XAPI:   AdapterXAPI{Username: "username1", Password: "password2", Tracker: "tracker3"},
				Syncer: &Syncer{Key: "override"}}},
		},
		{
			description: "Override XAPI",
			givenFsBidderInfos: BidderInfos{"a": {
				XAPI: AdapterXAPI{Username: "username", Password: "password", Tracker: "tracker"}}},
			givenConfigBidderInfos: BidderInfos{"a": {
				XAPI:   AdapterXAPI{Username: "username1", Password: "password2", Tracker: "tracker3"},
				Syncer: &Syncer{Key: "override"}}},
			expectedBidderInfos: BidderInfos{"a": {
				XAPI:   AdapterXAPI{Username: "username1", Password: "password2", Tracker: "tracker3"},
				Syncer: &Syncer{Key: "override"}}},
		},
		{
			description:            "Don't override PlatformID",
			givenFsBidderInfos:     BidderInfos{"a": {PlatformID: "PlatformID"}},
			givenConfigBidderInfos: BidderInfos{"a": {Syncer: &Syncer{Key: "override"}}},
			expectedBidderInfos:    BidderInfos{"a": {PlatformID: "PlatformID", Syncer: &Syncer{Key: "override"}}},
		},
		{
			description:            "Override PlatformID",
			givenFsBidderInfos:     BidderInfos{"a": {PlatformID: "PlatformID1"}},
			givenConfigBidderInfos: BidderInfos{"a": {PlatformID: "PlatformID2", Syncer: &Syncer{Key: "override"}}},
			expectedBidderInfos:    BidderInfos{"a": {PlatformID: "PlatformID2", Syncer: &Syncer{Key: "override"}}},
		},
		{
			description:            "Don't override AppSecret",
			givenFsBidderInfos:     BidderInfos{"a": {AppSecret: "AppSecret"}},
			givenConfigBidderInfos: BidderInfos{"a": {Syncer: &Syncer{Key: "override"}}},
			expectedBidderInfos:    BidderInfos{"a": {AppSecret: "AppSecret", Syncer: &Syncer{Key: "override"}}},
		},
		{
			description:            "Override AppSecret",
			givenFsBidderInfos:     BidderInfos{"a": {AppSecret: "AppSecret1"}},
			givenConfigBidderInfos: BidderInfos{"a": {AppSecret: "AppSecret2", Syncer: &Syncer{Key: "override"}}},
			expectedBidderInfos:    BidderInfos{"a": {AppSecret: "AppSecret2", Syncer: &Syncer{Key: "override"}}},
		},
	}
	for _, test := range testCases {
		bidderInfos, resultErr := applyBidderInfoConfigOverrides(test.givenConfigBidderInfos, test.givenFsBidderInfos)
		assert.NoError(t, resultErr, test.description+":err")
		assert.Equal(t, test.expectedBidderInfos, bidderInfos, test.description+":result")
	}
}
