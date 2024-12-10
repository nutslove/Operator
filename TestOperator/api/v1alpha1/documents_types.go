/*
Copyright 2024.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// DocumentsSpec defines the desired state of Documents
type DocumentsSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Gitリポジトリ(owner/repo形式)
	Repo string `json:"repo,omitempty"`
	// 監視対象のブランチ(デフォルト: main等)
	Branch string `json:"branch,omitempty"`

	// 複数の特定ファイルを監視したい場合はここで指定
	// 例: ["docs/file1.md", "docs/file2.md"]
	FilePaths []string `json:"filePaths,omitempty"`

	// ディレクトリを指定した場合、そのディレクトリ内の全てのmdファイルが対象
	// 例: "docs/"
	Directory string `json:"directory,omitempty"`

	// 除外パターン: 正規表現でmdファイルを除外可能にする
	// 例: ["^docs/ignore_.*\\.md$", ".*draft\\.md$"]
	ExcludePatterns []string `json:"excludePatterns,omitempty"`

	// ポーリング間隔(秒)
	IntervalSeconds int `json:"intervalSeconds,omitempty"`
}

// DocumentsStatus defines the observed state of Documents
type DocumentsStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	// ファイルパス (リポジトリ基準パス)
	FilePath string `json:"filePath,omitempty"`
	// 最終確認時のSHAハッシュ
	LastKnownSha string `json:"lastKnownSha,omitempty"`
}

type AllDocumentsStatus struct {
	// 複数ファイル分のSHAを記録する
	Documents []DocumentsStatus `json:"Documents,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// Documents is the Schema for the documents API
type Documents struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DocumentsSpec   `json:"spec,omitempty"`
	Status DocumentsStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// DocumentsList contains a list of Documents
type DocumentsList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Documents `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Documents{}, &DocumentsList{})
}
