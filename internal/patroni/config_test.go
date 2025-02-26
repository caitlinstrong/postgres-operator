// Copyright 2021 - 2025 Crunchy Data Solutions, Inc.
//
// SPDX-License-Identifier: Apache-2.0

package patroni

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"gotest.tools/v3/assert"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/yaml"

	"github.com/crunchydata/postgres-operator/internal/postgres"
	"github.com/crunchydata/postgres-operator/internal/testing/cmp"
	"github.com/crunchydata/postgres-operator/internal/testing/require"
	"github.com/crunchydata/postgres-operator/pkg/apis/postgres-operator.crunchydata.com/v1beta1"
)

func TestClusterYAML(t *testing.T) {
	t.Parallel()

	t.Run("PG version defaulted", func(t *testing.T) {
		cluster := new(v1beta1.PostgresCluster)
		cluster.Default()
		cluster.Namespace = "some-namespace"
		cluster.Name = "cluster-name"

		data, err := clusterYAML(cluster, postgres.HBAs{}, postgres.Parameters{}, 0)
		assert.NilError(t, err)
		assert.Equal(t, data, strings.TrimSpace(`
# Generated by postgres-operator. DO NOT EDIT.
# Your changes will not be saved.
bootstrap:
  dcs:
    loop_wait: 10
    postgresql:
      parameters: {}
      pg_hba: []
      use_pg_rewind: false
      use_slots: false
    ttl: 30
ctl:
  cacert: /etc/patroni/~postgres-operator/patroni.ca-roots
  certfile: /etc/patroni/~postgres-operator/patroni.crt+key
  insecure: false
  keyfile: null
kubernetes:
  labels:
    postgres-operator.crunchydata.com/cluster: cluster-name
  leader_label_value: master
  namespace: some-namespace
  role_label: postgres-operator.crunchydata.com/role
  scope_label: postgres-operator.crunchydata.com/patroni
  use_endpoints: true
postgresql:
  authentication:
    replication:
      sslcert: /tmp/replication/tls.crt
      sslkey: /tmp/replication/tls.key
      sslmode: verify-ca
      sslrootcert: /tmp/replication/ca.crt
      username: _crunchyrepl
    rewind:
      sslcert: /tmp/replication/tls.crt
      sslkey: /tmp/replication/tls.key
      sslmode: verify-ca
      sslrootcert: /tmp/replication/ca.crt
      username: _crunchyrepl
restapi:
  cafile: /etc/patroni/~postgres-operator/patroni.ca-roots
  certfile: /etc/patroni/~postgres-operator/patroni.crt+key
  keyfile: null
  verify_client: optional
scope: cluster-name-ha
watchdog:
  mode: "off"
	`)+"\n")
	})

	t.Run(">PG10", func(t *testing.T) {
		cluster := new(v1beta1.PostgresCluster)
		cluster.Default()
		cluster.Namespace = "some-namespace"
		cluster.Name = "cluster-name"
		cluster.Spec.PostgresVersion = 14

		data, err := clusterYAML(cluster, postgres.HBAs{}, postgres.Parameters{}, 0)
		assert.NilError(t, err)
		assert.Equal(t, data, strings.TrimSpace(`
# Generated by postgres-operator. DO NOT EDIT.
# Your changes will not be saved.
bootstrap:
  dcs:
    loop_wait: 10
    postgresql:
      parameters: {}
      pg_hba: []
      use_pg_rewind: true
      use_slots: false
    ttl: 30
ctl:
  cacert: /etc/patroni/~postgres-operator/patroni.ca-roots
  certfile: /etc/patroni/~postgres-operator/patroni.crt+key
  insecure: false
  keyfile: null
kubernetes:
  labels:
    postgres-operator.crunchydata.com/cluster: cluster-name
  leader_label_value: master
  namespace: some-namespace
  role_label: postgres-operator.crunchydata.com/role
  scope_label: postgres-operator.crunchydata.com/patroni
  use_endpoints: true
postgresql:
  authentication:
    replication:
      sslcert: /tmp/replication/tls.crt
      sslkey: /tmp/replication/tls.key
      sslmode: verify-ca
      sslrootcert: /tmp/replication/ca.crt
      username: _crunchyrepl
    rewind:
      sslcert: /tmp/replication/tls.crt
      sslkey: /tmp/replication/tls.key
      sslmode: verify-ca
      sslrootcert: /tmp/replication/ca.crt
      username: _crunchyrepl
restapi:
  cafile: /etc/patroni/~postgres-operator/patroni.ca-roots
  certfile: /etc/patroni/~postgres-operator/patroni.crt+key
  keyfile: null
  verify_client: optional
scope: cluster-name-ha
watchdog:
  mode: "off"
	`)+"\n")
	})

	t.Run("PatroniLogSizeConfigured", func(t *testing.T) {
		cluster := new(v1beta1.PostgresCluster)
		cluster.Default()
		cluster.Namespace = "some-namespace"
		cluster.Name = "cluster-name"
		cluster.Spec.PostgresVersion = 14

		fileSize, err := resource.ParseQuantity("1k")
		assert.NilError(t, err)

		logLevel := "DEBUG"
		cluster.Spec.Patroni.Logging = &v1beta1.PatroniLogConfig{
			StorageLimit: &fileSize,
			Level:        &logLevel,
		}

		data, err := clusterYAML(cluster, postgres.HBAs{}, postgres.Parameters{}, 1000)
		assert.NilError(t, err)
		assert.Equal(t, data, strings.TrimSpace(`
# Generated by postgres-operator. DO NOT EDIT.
# Your changes will not be saved.
bootstrap:
  dcs:
    loop_wait: 10
    postgresql:
      parameters: {}
      pg_hba: []
      use_pg_rewind: true
      use_slots: false
    ttl: 30
ctl:
  cacert: /etc/patroni/~postgres-operator/patroni.ca-roots
  certfile: /etc/patroni/~postgres-operator/patroni.crt+key
  insecure: false
  keyfile: null
kubernetes:
  labels:
    postgres-operator.crunchydata.com/cluster: cluster-name
  leader_label_value: master
  namespace: some-namespace
  role_label: postgres-operator.crunchydata.com/role
  scope_label: postgres-operator.crunchydata.com/patroni
  use_endpoints: true
log:
  dir: /pgdata/patroni/log
  file_num: 1
  file_size: 500
  level: DEBUG
  mode: "0660"
  type: json
postgresql:
  authentication:
    replication:
      sslcert: /tmp/replication/tls.crt
      sslkey: /tmp/replication/tls.key
      sslmode: verify-ca
      sslrootcert: /tmp/replication/ca.crt
      username: _crunchyrepl
    rewind:
      sslcert: /tmp/replication/tls.crt
      sslkey: /tmp/replication/tls.key
      sslmode: verify-ca
      sslrootcert: /tmp/replication/ca.crt
      username: _crunchyrepl
restapi:
  cafile: /etc/patroni/~postgres-operator/patroni.ca-roots
  certfile: /etc/patroni/~postgres-operator/patroni.crt+key
  keyfile: null
  verify_client: optional
scope: cluster-name-ha
watchdog:
  mode: "off"
	`)+"\n")
	})
}

func TestDynamicConfiguration(t *testing.T) {
	t.Parallel()

	parameters := func(in map[string]string) *postgres.ParameterSet {
		out := postgres.NewParameterSet()
		for k, v := range in {
			out.Add(k, v)
		}
		return out
	}

	for _, tt := range []struct {
		name     string
		spec     string
		hbas     postgres.HBAs
		params   postgres.Parameters
		expected map[string]any
	}{
		{
			name: "empty is valid",
			expected: map[string]any{
				"loop_wait": int32(10),
				"ttl":       int32(30),
				"postgresql": map[string]any{
					"parameters":    map[string]any{},
					"pg_hba":        []string{},
					"use_pg_rewind": true,
					"use_slots":     false,
				},
			},
		},
		{
			name: "top-level passes through",
			spec: `{
				patroni: {
					dynamicConfiguration: {
						retry_timeout: 5,
					},
				},
			}`,
			expected: map[string]any{
				"loop_wait":     int32(10),
				"ttl":           int32(30),
				"retry_timeout": int64(5),
				"postgresql": map[string]any{
					"parameters":    map[string]any{},
					"pg_hba":        []string{},
					"use_pg_rewind": true,
					"use_slots":     false,
				},
			},
		},
		{
			name: "top-level: spec overrides input",
			spec: `{
				patroni: {
					leaderLeaseDurationSeconds: 99,
					syncPeriodSeconds: 8,
					dynamicConfiguration: {
						loop_wait: 3,
						ttl: nope,
					},
				},
			}`,
			expected: map[string]any{
				"loop_wait": int32(8),
				"ttl":       int32(99),
				"postgresql": map[string]any{
					"parameters":    map[string]any{},
					"pg_hba":        []string{},
					"use_pg_rewind": true,
					"use_slots":     false,
				},
			},
		},
		{
			name: "postgresql: wrong-type is ignored",
			spec: `{
				patroni: {
					dynamicConfiguration: {
						postgresql: true,
					},
				},
			}`,
			expected: map[string]any{
				"loop_wait": int32(10),
				"ttl":       int32(30),
				"postgresql": map[string]any{
					"parameters":    map[string]any{},
					"pg_hba":        []string{},
					"use_pg_rewind": true,
					"use_slots":     false,
				},
			},
		},
		{
			name: "postgresql: defaults and overrides",
			spec: `{
				patroni: {
					dynamicConfiguration: {
						postgresql: {
							use_pg_rewind: overidden,
							use_slots: input,
						},
					},
				},
			}`,
			expected: map[string]any{
				"loop_wait": int32(10),
				"ttl":       int32(30),
				"postgresql": map[string]any{
					"parameters":    map[string]any{},
					"pg_hba":        []string{},
					"use_pg_rewind": true,
					"use_slots":     "input",
				},
			},
		},
		{
			name: "postgresql.parameters: wrong-type is ignored",
			spec: `{
				patroni: {
					dynamicConfiguration: {
						postgresql: {
							parameters: true,
						},
					},
				},
			}`,
			expected: map[string]any{
				"loop_wait": int32(10),
				"ttl":       int32(30),
				"postgresql": map[string]any{
					"parameters":    map[string]any{},
					"pg_hba":        []string{},
					"use_pg_rewind": true,
					"use_slots":     false,
				},
			},
		},
		{
			name: "postgresql.parameters: input passes through",
			spec: `{
				patroni: {
					dynamicConfiguration: {
						postgresql: {
							parameters: {
								something: str,
								another: 5,
							},
						},
					},
				},
			}`,
			expected: map[string]any{
				"loop_wait": int32(10),
				"ttl":       int32(30),
				"postgresql": map[string]any{
					"parameters": map[string]any{
						"something": "str",
						"another":   int64(5),
					},
					"pg_hba":        []string{},
					"use_pg_rewind": true,
					"use_slots":     false,
				},
			},
		},
		{
			name: "config.parameters takes precedence",
			spec: `{
				config: {
					parameters: {
						something: this,
					},
				},
				patroni: {
					dynamicConfiguration: {
						postgresql: {
							parameters: {
								something: str,
								another: 5,
							},
						},
					},
				},
			}`,
			expected: map[string]any{
				"loop_wait": int32(10),
				"ttl":       int32(30),
				"postgresql": map[string]any{
					"parameters": map[string]any{
						"something": intstr.FromString("this"),
						"another":   int64(5),
					},
					"pg_hba":        []string{},
					"use_pg_rewind": true,
					"use_slots":     false,
				},
			},
		},
		{
			name: "config.parameters: input overrides default",
			spec: `{
				config: {
					parameters: {
						something: str,
						another: 5,
					},
				},
			}`,
			params: postgres.Parameters{
				Default: parameters(map[string]string{
					"something": "overridden",
					"unrelated": "default",
				}),
			},
			expected: map[string]any{
				"loop_wait": int32(10),
				"ttl":       int32(30),
				"postgresql": map[string]any{
					"parameters": map[string]any{
						"something": intstr.FromString("str"),
						"another":   intstr.FromInt(5),
						"unrelated": "default",
					},
					"pg_hba":        []string{},
					"use_pg_rewind": true,
					"use_slots":     false,
				},
			},
		},
		{
			name: "config.parameters: mandatory overrides input",
			spec: `{
				config: {
					parameters: {
						something: str,
						another: 5,
					},
				},
			}`,
			params: postgres.Parameters{
				Mandatory: parameters(map[string]string{
					"something": "overrides",
					"unrelated": "setting",
				}),
			},
			expected: map[string]any{
				"loop_wait": int32(10),
				"ttl":       int32(30),
				"postgresql": map[string]any{
					"parameters": map[string]any{
						"something": "overrides",
						"another":   intstr.FromInt(5),
						"unrelated": "setting",
					},
					"pg_hba":        []string{},
					"use_pg_rewind": true,
					"use_slots":     false,
				},
			},
		},
		{
			name: "config.parameters: mandatory shared_preload_libraries",
			spec: `{
				config: {
					parameters: {
						shared_preload_libraries: given,
					},
				},
			}`,
			params: postgres.Parameters{
				Mandatory: parameters(map[string]string{
					"shared_preload_libraries": "mandatory",
				}),
			},
			expected: map[string]any{
				"loop_wait": int32(10),
				"ttl":       int32(30),
				"postgresql": map[string]any{
					"parameters": map[string]any{
						"shared_preload_libraries": "mandatory,given",
					},
					"pg_hba":        []string{},
					"use_pg_rewind": true,
					"use_slots":     false,
				},
			},
		},
		{
			name: "config.parameters: mandatory shared_preload_libraries wrong-type is ignored",
			spec: `{
				config: {
					parameters: {
						shared_preload_libraries: 1,
					},
				},
			}`,
			params: postgres.Parameters{
				Mandatory: parameters(map[string]string{
					"shared_preload_libraries": "mandatory",
				}),
			},
			expected: map[string]any{
				"loop_wait": int32(10),
				"ttl":       int32(30),
				"postgresql": map[string]any{
					"parameters": map[string]any{
						"shared_preload_libraries": "mandatory",
					},
					"pg_hba":        []string{},
					"use_pg_rewind": true,
					"use_slots":     false,
				},
			},
		},
		{
			name: "config.parameters: shared_preload_libraries order",
			spec: `{
				config: {
					parameters: {
						shared_preload_libraries: "given, citus, more",
					},
				},
			}`,
			params: postgres.Parameters{
				Mandatory: parameters(map[string]string{
					"shared_preload_libraries": "mandatory",
				}),
			},
			expected: map[string]any{
				"loop_wait": int32(10),
				"ttl":       int32(30),
				"postgresql": map[string]any{
					"parameters": map[string]any{
						"shared_preload_libraries": "citus,mandatory,given, citus, more",
					},
					"pg_hba":        []string{},
					"use_pg_rewind": true,
					"use_slots":     false,
				},
			},
		},
		{
			name: "postgresql.pg_hba: wrong-type is ignored",
			spec: `{
				patroni: {
					dynamicConfiguration: {
						postgresql: {
							pg_hba: true,
						},
					},
				},
			}`,
			expected: map[string]any{
				"loop_wait": int32(10),
				"ttl":       int32(30),
				"postgresql": map[string]any{
					"parameters":    map[string]any{},
					"pg_hba":        []string{},
					"use_pg_rewind": true,
					"use_slots":     false,
				},
			},
		},
		{
			name: "postgresql.pg_hba: default when no input",
			spec: `{
				patroni: {
					dynamicConfiguration: {
						postgresql: {
							pg_hba: null,
						},
					},
				},
			}`,
			hbas: postgres.HBAs{
				Default: []*postgres.HostBasedAuthentication{
					postgres.NewHBA().Local().Method("peer"),
				},
			},
			expected: map[string]any{
				"loop_wait": int32(10),
				"ttl":       int32(30),
				"postgresql": map[string]any{
					"parameters": map[string]any{},
					"pg_hba": []string{
						"local all all peer",
					},
					"use_pg_rewind": true,
					"use_slots":     false,
				},
			},
		},
		{
			name: "postgresql.pg_hba: no default when input",
			spec: `{
				patroni: {
					dynamicConfiguration: {
						postgresql: {
							pg_hba: [custom],
						},
					},
				},
			}`,
			hbas: postgres.HBAs{
				Default: []*postgres.HostBasedAuthentication{
					postgres.NewHBA().Local().Method("peer"),
				},
			},
			expected: map[string]any{
				"loop_wait": int32(10),
				"ttl":       int32(30),
				"postgresql": map[string]any{
					"parameters": map[string]any{},
					"pg_hba": []string{
						"custom",
					},
					"use_pg_rewind": true,
					"use_slots":     false,
				},
			},
		},
		{
			name: "postgresql.pg_hba: mandatory before others",
			spec: `{
				patroni: {
					dynamicConfiguration: {
						postgresql: {
							pg_hba: [custom],
						},
					},
				},
			}`,
			hbas: postgres.HBAs{
				Mandatory: []*postgres.HostBasedAuthentication{
					postgres.NewHBA().Local().Method("peer"),
				},
			},
			expected: map[string]any{
				"loop_wait": int32(10),
				"ttl":       int32(30),
				"postgresql": map[string]any{
					"parameters": map[string]any{},
					"pg_hba": []string{
						"local all all peer",
						"custom",
					},
					"use_pg_rewind": true,
					"use_slots":     false,
				},
			},
		},
		{
			name: "postgresql.pg_hba: ignore non-string types",
			spec: `{
				patroni: {
					dynamicConfiguration: {
						postgresql: {
							pg_hba: [1, true, custom, {}, []],
						},
					},
				},
			}`,
			hbas: postgres.HBAs{
				Mandatory: []*postgres.HostBasedAuthentication{
					postgres.NewHBA().Local().Method("peer"),
				},
			},
			expected: map[string]any{
				"loop_wait": int32(10),
				"ttl":       int32(30),
				"postgresql": map[string]any{
					"parameters": map[string]any{},
					"pg_hba": []string{
						"local all all peer",
						"custom",
					},
					"use_pg_rewind": true,
					"use_slots":     false,
				},
			},
		},
		{
			name: "standby_cluster: input passes through",
			spec: `{
				patroni: {
					dynamicConfiguration: {
						standby_cluster: {
							primary_slot_name: str,
						},
					},
				},
			}`,
			expected: map[string]any{
				"loop_wait": int32(10),
				"ttl":       int32(30),
				"postgresql": map[string]any{
					"parameters":    map[string]any{},
					"pg_hba":        []string{},
					"use_pg_rewind": true,
					"use_slots":     false,
				},
				"standby_cluster": map[string]any{
					"primary_slot_name": "str",
				},
			},
		},
		{
			name: "standby_cluster: repo only",
			spec: `{
				standby: {
					enabled: true,
					repoName: repo,
				},
				patroni: {
					dynamicConfiguration: {
						standby_cluster: {
							restore_command: overridden,
							unrelated: input,
						},
					},
				},
			}`,
			params: postgres.Parameters{
				Mandatory: parameters(map[string]string{
					"restore_command": "mandatory",
				}),
			},
			expected: map[string]any{
				"loop_wait": int32(10),
				"ttl":       int32(30),
				"postgresql": map[string]any{
					"parameters": map[string]any{
						"restore_command": "mandatory",
					},
					"pg_hba":        []string{},
					"use_pg_rewind": true,
					"use_slots":     false,
				},
				"standby_cluster": map[string]any{
					"create_replica_methods": []string{"pgbackrest"},
					"restore_command":        "mandatory",
					"unrelated":              "input",
				},
			},
		},
		{
			name: "standby_cluster: basebackup for streaming",
			spec: `{
				standby: {
					enabled: true,
					host: 0.0.0.0,
					port: 5432,
				},
				patroni: {
					dynamicConfiguration: {
						standby_cluster: {
							host: overridden,
							port: 0000,
							restore_command: overridden,
							unrelated: input,
						},
					},
				},
			}`,
			params: postgres.Parameters{
				Mandatory: parameters(map[string]string{
					"restore_command": "mandatory",
				}),
			},
			expected: map[string]any{
				"loop_wait": int32(10),
				"ttl":       int32(30),
				"postgresql": map[string]any{
					"parameters": map[string]any{
						"restore_command": "mandatory",
					},
					"pg_hba":        []string{},
					"use_pg_rewind": true,
					"use_slots":     false,
				},
				"standby_cluster": map[string]any{
					"create_replica_methods": []string{"basebackup"},
					"host":                   "0.0.0.0",
					"port":                   int32(5432),
					"unrelated":              "input",
				},
			},
		},
		{
			name: "standby_cluster: both repo and streaming",
			spec: `{
				standby: {
					enabled: true,
					host: 0.0.0.0,
					port: 5432,
					repoName: repo,
				},
				patroni: {
					dynamicConfiguration: {
						standby_cluster: {
							host: overridden,
							port: 9999,
							restore_command: overridden,
							unrelated: input,
						},
					},
				},
			}`,
			params: postgres.Parameters{
				Mandatory: parameters(map[string]string{
					"restore_command": "mandatory",
				}),
			},
			expected: map[string]any{
				"loop_wait": int32(10),
				"ttl":       int32(30),
				"postgresql": map[string]any{
					"parameters": map[string]any{
						"restore_command": "mandatory",
					},
					"pg_hba":        []string{},
					"use_pg_rewind": true,
					"use_slots":     false,
				},
				"standby_cluster": map[string]any{
					"create_replica_methods": []string{"pgbackrest", "basebackup"},
					"host":                   "0.0.0.0",
					"port":                   int32(5432),
					"restore_command":        "mandatory",
					"unrelated":              "input",
				},
			},
		},
		{
			name: "config.parameters: tde enabled",
			spec: `{
				config: {
					parameters: {
						encryption_key_command: echo one,
					},
				},
			}`,
			expected: map[string]any{
				"loop_wait": int32(10),
				"ttl":       int32(30),
				"postgresql": map[string]any{
					"bin_name": map[string]any{"pg_rewind": string("/tmp/pg_rewind_tde.sh")},
					"parameters": map[string]any{
						"encryption_key_command": intstr.FromString("echo one"),
					},
					"pg_hba":        []string{},
					"use_pg_rewind": bool(true),
					"use_slots":     bool(false),
				},
			},
		},
		{
			name: "postgresql.parameters: tde enabled",
			spec: `{
				patroni: {
					dynamicConfiguration: {
						postgresql: {
							parameters: {
								encryption_key_command: echo test,
							},
						},
					},
				},
			}`,
			expected: map[string]any{
				"loop_wait": int32(10),
				"ttl":       int32(30),
				"postgresql": map[string]any{
					"bin_name": map[string]any{"pg_rewind": string("/tmp/pg_rewind_tde.sh")},
					"parameters": map[string]any{
						"encryption_key_command": "echo test",
					},
					"pg_hba":        []string{},
					"use_pg_rewind": bool(true),
					"use_slots":     bool(false),
				},
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			cluster := new(v1beta1.PostgresCluster)
			require.UnmarshalInto(t, &cluster.Spec, tt.spec)
			if cluster.Spec.PostgresVersion == 0 {
				cluster.Spec.PostgresVersion = 14
			}
			cluster.Default()
			actual := DynamicConfiguration(&cluster.Spec, tt.hbas, tt.params)
			assert.DeepEqual(t, tt.expected, actual)
		})
	}
}

func TestInstanceConfigFiles(t *testing.T) {
	t.Parallel()

	cm1 := &corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: "cm1"}}
	cm2 := &corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: "cm2"}}

	projections := instanceConfigFiles(cm1, cm2)

	assert.Assert(t, cmp.MarshalMatches(projections, `
- configMap:
    items:
    - key: patroni.yaml
      path: ~postgres-operator_cluster.yaml
    name: cm1
- configMap:
    items:
    - key: patroni.yaml
      path: ~postgres-operator_instance.yaml
    name: cm2
	`))
}

func TestInstanceEnvironment(t *testing.T) {
	t.Parallel()

	cluster := new(v1beta1.PostgresCluster)
	cluster.Default()
	cluster.Spec.PostgresVersion = 12
	leaderService := new(corev1.Service)
	podService := new(corev1.Service)
	podService.Name = "pod-dns"

	vars := instanceEnvironment(cluster, podService, leaderService, nil)

	assert.Assert(t, cmp.MarshalMatches(vars, `
- name: PATRONI_NAME
  valueFrom:
    fieldRef:
      apiVersion: v1
      fieldPath: metadata.name
- name: PATRONI_KUBERNETES_POD_IP
  valueFrom:
    fieldRef:
      apiVersion: v1
      fieldPath: status.podIP
- name: PATRONI_KUBERNETES_PORTS
  value: |
    []
- name: PATRONI_POSTGRESQL_CONNECT_ADDRESS
  value: $(PATRONI_NAME).pod-dns:5432
- name: PATRONI_POSTGRESQL_LISTEN
  value: '*:5432'
- name: PATRONI_POSTGRESQL_CONFIG_DIR
  value: /pgdata/pg12
- name: PATRONI_POSTGRESQL_DATA_DIR
  value: /pgdata/pg12
- name: PATRONI_RESTAPI_CONNECT_ADDRESS
  value: $(PATRONI_NAME).pod-dns:8008
- name: PATRONI_RESTAPI_LISTEN
  value: '*:8008'
- name: PATRONICTL_CONFIG_FILE
  value: /etc/patroni
	`))

	t.Run("MatchingPorts", func(t *testing.T) {
		leaderService.Spec.Ports = []corev1.ServicePort{{Name: "postgres"}}
		leaderService.Spec.Ports[0].TargetPort.StrVal = "postgres"
		containers := []corev1.Container{{Name: "okay"}}
		containers[0].Ports = []corev1.ContainerPort{{
			Name: "postgres", ContainerPort: 9999, Protocol: corev1.ProtocolTCP,
		}}

		vars := instanceEnvironment(cluster, podService, leaderService, containers)

		assert.Assert(t, cmp.MarshalMatches(vars, `
- name: PATRONI_NAME
  valueFrom:
    fieldRef:
      apiVersion: v1
      fieldPath: metadata.name
- name: PATRONI_KUBERNETES_POD_IP
  valueFrom:
    fieldRef:
      apiVersion: v1
      fieldPath: status.podIP
- name: PATRONI_KUBERNETES_PORTS
  value: |
    - name: postgres
      port: 9999
      protocol: TCP
- name: PATRONI_POSTGRESQL_CONNECT_ADDRESS
  value: $(PATRONI_NAME).pod-dns:5432
- name: PATRONI_POSTGRESQL_LISTEN
  value: '*:5432'
- name: PATRONI_POSTGRESQL_CONFIG_DIR
  value: /pgdata/pg12
- name: PATRONI_POSTGRESQL_DATA_DIR
  value: /pgdata/pg12
- name: PATRONI_RESTAPI_CONNECT_ADDRESS
  value: $(PATRONI_NAME).pod-dns:8008
- name: PATRONI_RESTAPI_LISTEN
  value: '*:8008'
- name: PATRONICTL_CONFIG_FILE
  value: /etc/patroni
		`))
	})
}

func TestInstanceYAML(t *testing.T) {
	t.Parallel()

	cluster := &v1beta1.PostgresCluster{Spec: v1beta1.PostgresClusterSpec{PostgresVersion: 12}}
	instance := new(v1beta1.PostgresInstanceSetSpec)

	data, err := instanceYAML(cluster, instance, nil)
	assert.NilError(t, err)
	assert.Equal(t, data, strings.Trim(`
# Generated by postgres-operator. DO NOT EDIT.
# Your changes will not be saved.
bootstrap:
  initdb:
  - data-checksums
  - encoding=UTF8
  - waldir=/pgdata/pg12_wal
  method: initdb
kubernetes: {}
postgresql:
  basebackup:
  - waldir=/pgdata/pg12_wal
  create_replica_methods:
  - basebackup
  pgpass: /tmp/.pgpass
  use_unix_socket: true
restapi: {}
tags: {}
	`, "\t\n")+"\n")

	dataWithReplicaCreate, err := instanceYAML(cluster, instance, []string{"some", "backrest", "cmd"})
	assert.NilError(t, err)
	assert.Equal(t, dataWithReplicaCreate, strings.Trim(`
# Generated by postgres-operator. DO NOT EDIT.
# Your changes will not be saved.
bootstrap:
  initdb:
  - data-checksums
  - encoding=UTF8
  - waldir=/pgdata/pg12_wal
  method: initdb
kubernetes: {}
postgresql:
  basebackup:
  - waldir=/pgdata/pg12_wal
  create_replica_methods:
  - pgbackrest
  - basebackup
  pgbackrest:
    command: '''bash'' ''-ceu'' ''--'' ''install --directory --mode=0700 "${PGDATA?}"
      && exec "$@"'' ''-'' ''some'' ''backrest'' ''cmd'''
    keep_data: true
    no_leader: true
    no_params: true
  pgpass: /tmp/.pgpass
  use_unix_socket: true
restapi: {}
tags: {}
	`, "\t\n")+"\n")

	cluster.Spec.Patroni = &v1beta1.PatroniSpec{
		DynamicConfiguration: map[string]any{
			"postgresql": map[string]any{
				"parameters": map[string]any{
					"encryption_key_command": "echo test",
				},
			},
		},
	}

	datawithTDE, err := instanceYAML(cluster, instance, nil)
	assert.NilError(t, err)
	assert.Equal(t, datawithTDE, strings.Trim(`
# Generated by postgres-operator. DO NOT EDIT.
# Your changes will not be saved.
bootstrap:
  initdb:
  - data-checksums
  - encoding=UTF8
  - waldir=/pgdata/pg12_wal
  - encryption-key-command=echo test
  method: initdb
kubernetes: {}
postgresql:
  basebackup:
  - waldir=/pgdata/pg12_wal
  create_replica_methods:
  - basebackup
  pgpass: /tmp/.pgpass
  use_unix_socket: true
restapi: {}
tags: {}
	`, "\t\n")+"\n")

}

func TestPGBackRestCreateReplicaCommand(t *testing.T) {
	t.Parallel()

	shellcheck := require.ShellCheck(t)
	cluster := new(v1beta1.PostgresCluster)
	instance := new(v1beta1.PostgresInstanceSetSpec)

	data, err := instanceYAML(cluster, instance, []string{"some", "backrest", "cmd"})
	assert.NilError(t, err)

	var parsed struct {
		PostgreSQL struct {
			PGBackRest struct {
				Command string
			}
		}
	}
	assert.NilError(t, yaml.Unmarshal([]byte(data), &parsed))

	dir := t.TempDir()

	// The command should be compatible with any shell.
	{
		command := parsed.PostgreSQL.PGBackRest.Command
		file := filepath.Join(dir, "command.sh")
		assert.NilError(t, os.WriteFile(file, []byte(command), 0o600))

		cmd := exec.Command(shellcheck, "--enable=all", "--shell=sh", file)
		output, err := cmd.CombinedOutput()
		assert.NilError(t, err, "%q\n%s", cmd.Args, output)
	}

	// Naive parsing of shell words...
	command := strings.Split(strings.Trim(parsed.PostgreSQL.PGBackRest.Command, "'"), "' '")

	// Expect a bash command with an inline script.
	assert.DeepEqual(t, command[:3], []string{"bash", "-ceu", "--"})
	assert.Assert(t, len(command) > 3)
	script := command[3]

	// It should call the pgBackRest command.
	assert.Assert(t, strings.HasSuffix(script, ` exec "$@"`))
	assert.DeepEqual(t, command[len(command)-3:], []string{"some", "backrest", "cmd"})

	// It should pass shellcheck.
	{
		file := filepath.Join(dir, "script.bash")
		assert.NilError(t, os.WriteFile(file, []byte(script), 0o600))

		cmd := exec.Command(shellcheck, "--enable=all", file)
		output, err := cmd.CombinedOutput()
		assert.NilError(t, err, "%q\n%s", cmd.Args, output)
	}
}

func TestProbeTiming(t *testing.T) {
	t.Parallel()

	defaults := new(v1beta1.PatroniSpec)
	defaults.Default()

	// Defaults should match the suggested/documented timing.
	// - https://github.com/zalando/patroni/blob/v2.0.1/docs/rest_api.rst
	assert.DeepEqual(t, probeTiming(defaults), &corev1.Probe{
		TimeoutSeconds:   5,
		PeriodSeconds:    10,
		SuccessThreshold: 1,
		FailureThreshold: 3,
	})

	for _, tt := range []struct {
		lease, sync int32
		expected    corev1.Probe
	}{
		// The smallest possible values for "loop_wait" and "retry_timeout" are
		// both 1 sec which makes 3 sec the smallest appropriate value for "ttl".
		// These are the validation minimums in v1beta1.PatroniSpec.
		{lease: 3, sync: 1, expected: corev1.Probe{
			TimeoutSeconds:   1,
			PeriodSeconds:    1,
			SuccessThreshold: 1,
			FailureThreshold: 3,
		}},

		// These are plausible values for "ttl" and "loop_wait".
		{lease: 60, sync: 15, expected: corev1.Probe{
			TimeoutSeconds:   7,
			PeriodSeconds:    15,
			SuccessThreshold: 1,
			FailureThreshold: 4,
		}},
		{lease: 10, sync: 5, expected: corev1.Probe{
			TimeoutSeconds:   2,
			PeriodSeconds:    5,
			SuccessThreshold: 1,
			FailureThreshold: 2,
		}},

		// These are plausible values that aren't multiples of each other.
		// Failure triggers sooner than "ttl", which seems to agree with docs:
		// - https://github.com/zalando/patroni/blob/v2.0.1/docs/watchdog.rst
		{lease: 19, sync: 7, expected: corev1.Probe{
			TimeoutSeconds:   3,
			PeriodSeconds:    7,
			SuccessThreshold: 1,
			FailureThreshold: 2,
		}},
		{lease: 13, sync: 7, expected: corev1.Probe{
			TimeoutSeconds:   3,
			PeriodSeconds:    7,
			SuccessThreshold: 1,
			FailureThreshold: 1,
		}},

		// These values are infeasible for Patroni but produce valid v1.Probes.
		{lease: 60, sync: 60, expected: corev1.Probe{
			TimeoutSeconds:   30,
			PeriodSeconds:    60,
			SuccessThreshold: 1,
			FailureThreshold: 1,
		}},
		{lease: 10, sync: 20, expected: corev1.Probe{
			TimeoutSeconds:   10,
			PeriodSeconds:    20,
			SuccessThreshold: 1,
			FailureThreshold: 1,
		}},
	} {
		tt := tt
		actual := probeTiming(&v1beta1.PatroniSpec{
			LeaderLeaseDurationSeconds: &tt.lease,
			SyncPeriodSeconds:          &tt.sync,
		})
		assert.DeepEqual(t, actual, &tt.expected)

		// v1.Probe validation
		assert.Assert(t, actual.TimeoutSeconds >= 1)   // Minimum value is 1.
		assert.Assert(t, actual.PeriodSeconds >= 1)    // Minimum value is 1.
		assert.Assert(t, actual.SuccessThreshold == 1) // Must be 1 for liveness and startup.
		assert.Assert(t, actual.FailureThreshold >= 1) // Minimum value is 1.
	}
}
