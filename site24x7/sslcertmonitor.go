package site24x7

import (
	"fmt"
	"sort"
	"strconv"

	site24x7 "github.com/Bonial-International-GmbH/site24x7-go"
	"github.com/Bonial-International-GmbH/site24x7-go/api"
	apierrors "github.com/Bonial-International-GmbH/site24x7-go/api/errors"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

var SSLCertMonitorSchema = map[string]*schema.Schema{
	"display_name": {
		Type:     schema.TypeString,
		Required: true,
	},
	"domain_name": {
		Type:     schema.TypeString,
		Required: true,
	},
	"expire_days": {
		Type:     schema.TypeInt,
		Required: true,
	},
	"protocol": {
		Type:     schema.TypeString,
		Required: true,
	},
	"port": {
		Type:     schema.TypeInt,
		Required: true,
	},
	"timeout": {
		Type:     schema.TypeInt,
		Optional: true,
		Default:  10,
	},
	"location_profile_id": {
		Type:     schema.TypeString,
		Optional: true,
		Computed: true,
	},
	"notification_profile_id": {
		Type:     schema.TypeString,
		Optional: true,
		Computed: true,
	},
	"threshold_profile_id": {
		Type:     schema.TypeString,
		Optional: true,
		Computed: true,
	},
	"monitor_groups": {
		Type: schema.TypeList,
		Elem: &schema.Schema{
			Type: schema.TypeString,
		},
		Optional: true,
	},
	"user_group_ids": {
		Type: schema.TypeList,
		Elem: &schema.Schema{
			Type: schema.TypeString,
		},
		Optional: true,
		Computed: true,
	},
	"actions": {
		Type:     schema.TypeMap,
		Optional: true,
		Elem:     schema.TypeString,
	},
}

func resourceSite24x7SSLCertMonitor() *schema.Resource {
	return &schema.Resource{
		Create: SSLCertMonitorCreate,
		Read:   SSLCertMonitorRead,
		Update: SSLCertMonitorUpdate,
		Delete: SSLCertMonitorDelete,
		Exists: SSLCertMonitorExists,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: SSLCertMonitorSchema,
	}
}

func SSLCertMonitorCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(site24x7.Client)

	SSLCertMonitor, err := resourceDataToSSLCertMonitor(d, client)
	if err != nil {
		return err
	}

	SSLCertMonitor, err = client.Monitors().Create(SSLCertMonitor)
	if err != nil {
		return err
	}

	d.SetId(SSLCertMonitor.MonitorID)

	return nil
}

func SSLCertMonitorRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(site24x7.Client)

	SSLCertMonitor, err := client.Monitors().Get(d.Id())
	if err != nil {
		return err
	}

	updateSSLCertMonitorResourceData(d, SSLCertMonitor)

	return nil
}

func SSLCertMonitorUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(site24x7.Client)

	SSLCertMonitor, err := resourceDataToSSLCertMonitor(d, client)
	if err != nil {
		return err
	}

	SSLCertMonitor, err = client.Monitors().Update(SSLCertMonitor)
	if err != nil {
		return err
	}

	d.SetId(SSLCertMonitor.MonitorID)

	return nil
}

func SSLCertMonitorDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(site24x7.Client)

	err := client.Monitors().Delete(d.Id())
	if apierrors.IsNotFound(err) {
		return nil
	}

	return err
}

func SSLCertMonitorExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	client := meta.(site24x7.Client)

	_, err := client.Monitors().Get(d.Id())
	if apierrors.IsNotFound(err) {
		return false, nil
	}

	if err != nil {
		return false, err
	}

	return true, nil
}

func resourceDataToSSLCertMonitor(d *schema.ResourceData, client site24x7.Client) (*api.Monitor, error) {
	var userGroupIDs []string
	for _, id := range d.Get("user_group_ids").([]interface{}) {
		userGroupIDs = append(userGroupIDs, id.(string))
	}

	var monitorGroups []string
	for _, group := range d.Get("monitor_groups").([]interface{}) {
		monitorGroups = append(monitorGroups, group.(string))
	}

	actionMap := d.Get("actions").(map[string]interface{})

	keys := make([]string, 0, len(actionMap))
	for k := range actionMap {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	actionRefs := make([]api.ActionRef, len(keys))
	for i, k := range keys {
		status, err := strconv.Atoi(k)
		if err != nil {
			return nil, err
		}

		actionRefs[i] = api.ActionRef{
			ActionID:  actionMap[k].(string),
			AlertType: api.Status(status),
		}
	}

	SSLCertMonitor := &api.Monitor{
		MonitorID:             d.Id(),
		DisplayName:           d.Get("display_name").(string),
		Type:                  "SSL_CERT",
		DomainName:            d.Get("domain_name").(string),
		ExpireDays:            d.Get("expire_days").(int),
		Protocol:              d.Get("protocol").(string),
		Port:                  d.Get("port").(int),
		Timeout:               d.Get("timeout").(int),
		LocationProfileID:     d.Get("location_profile_id").(string),
		NotificationProfileID: d.Get("notification_profile_id").(string),
		ThresholdProfileID:    d.Get("threshold_profile_id").(string),
		MonitorGroups:         monitorGroups,
		UserGroupIDs:          userGroupIDs,
		ActionIDs:             actionRefs,
	}

	if SSLCertMonitor.LocationProfileID == "" {
		profile, err := DefaultLocationProfile(client)
		if err != nil {
			return nil, err
		}
		SSLCertMonitor.LocationProfileID = profile.ProfileID
		d.Set("location_profile_id", profile.ProfileID)
	}

	if SSLCertMonitor.NotificationProfileID == "" {
		profile, err := DefaultNotificationProfile(client)
		if err != nil {
			return nil, err
		}
		SSLCertMonitor.NotificationProfileID = profile.ProfileID
		d.Set("notification_profile_id", profile.ProfileID)
	}

	if SSLCertMonitor.ThresholdProfileID == "" {
		profile, err := DefaultThresholdProfile(client)
		if err != nil {
			return nil, err
		}
		SSLCertMonitor.ThresholdProfileID = profile.ProfileID
		d.Set("threshold_profile_id", profile)
	}

	if len(SSLCertMonitor.UserGroupIDs) == 0 {
		userGroup, err := DefaultUserGroup(client)
		if err != nil {
			return nil, err
		}
		SSLCertMonitor.UserGroupIDs = []string{userGroup.UserGroupID}
		d.Set("user_group_ids", []string{userGroup.UserGroupID})
	}

	return SSLCertMonitor, nil
}

func updateSSLCertMonitorResourceData(d *schema.ResourceData, monitor *api.Monitor) {
	d.Set("display_name", monitor.DisplayName)
	d.Set("domain_name", monitor.DomainName)
	d.Set("expire_days", monitor.ExpireDays)
	d.Set("protocol", monitor.Protocol)
	d.Set("port", monitor.Port)
	d.Set("timeout", monitor.Timeout)
	d.Set("location_profile_id", monitor.LocationProfileID)
	d.Set("notification_profile_id", monitor.NotificationProfileID)
	d.Set("threshold_profile_id", monitor.ThresholdProfileID)
	d.Set("monitor_groups", monitor.MonitorGroups)
	d.Set("user_group_ids", monitor.UserGroupIDs)

	actions := make(map[string]interface{})
	for _, action := range monitor.ActionIDs {
		actions[fmt.Sprintf("%d", action.AlertType)] = action.ActionID
	}

	d.Set("actions", actions)
}
