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

package controller

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	ragv1alpha1 "github.com/nutslove/Operator/api/v1alpha1"
)

// DocumentsReconciler reconciles a Documents object
type DocumentsReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=rag.nutslove,resources=documents,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=rag.nutslove,resources=documents/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=rag.nutslove,resources=documents/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Documents object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.18.4/pkg/reconcile
func (r *DocumentsReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	// TODO(user): your logic here

	var doc ragv1alpha1.Documents
	if err := r.Get(ctx, req.NamespacedName, &doc); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// GitHubファイルのSHA取得 (GitHub Personal Access Tokenが必要な場合はSecret参照)
	currentSha, err := fetchGitHubFileSha("nutslove", doc.Spec.Repo, doc.Spec.Branch, doc.Spec.FilePath, os.Getenv("GITHUB_TOKEN"))
	if err != nil {
		return ctrl.Result{}, err
	}

	if doc.Status.LastKnownSha == "" {
		// 初回記録
		doc.Status.LastKnownSha = currentSha
		if err := r.Status().Update(ctx, &doc); err != nil {
			return ctrl.Result{}, err
		}
	} else if doc.Status.LastKnownSha != currentSha {
		// // SHAが変わった場合、Python実行Jobを作成
		// err = r.createPythonJob(ctx, &doc)
		// if err != nil {
		// 	return ctrl.Result{}, err
		// }
		fmt.Println("Sha of Document has been changed.")

		// Status更新
		doc.Status.LastKnownSha = currentSha
		if err := r.Status().Update(ctx, &doc); err != nil {
			return ctrl.Result{}, err
		}
	}

	// 次回チェックまで待機
	return ctrl.Result{RequeueAfter: time.Duration(doc.Spec.IntervalSeconds) * time.Second}, nil

	// return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *DocumentsReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&ragv1alpha1.Documents{}).
		Complete(r)
}

// fetchGitHubFileSha はGitHubリポジトリ内の特定ファイルのSHAを取得する関数です。
// owner: GitHubリポジトリのオーナー名（例: "octocat"）
// repo: リポジトリ名（例: "Hello-World"）
// branch: 取得するブランチ（例: "main"）
// filePath: ファイルパス（例: "docs/example.md"）
// token: GitHub Personal Access Token（パブリックリポジトリで読み取りのみであれば不要な場合もありますが、
//
//	制限があるためトークンを用いておくことを推奨）
//
// 戻り値: ファイルのSHA文字列とエラー
func fetchGitHubFileSha(owner, repo, branch, filePath, token string) (string, error) {
	ctx := context.Background()

	var tc *http.Client
	if token != "" {
		ts := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: token},
		)
		tc = oauth2.NewClient(ctx, ts)
	} else {
		// トークンなしクライアント（レートリミットやプライベートリポジトリへのアクセスに制限あり）
		tc = http.DefaultClient
	}

	client := github.NewClient(tc)

	// GitHub APIでファイルコンテンツを取得
	fileContent, _, resp, err := client.Repositories.GetContents(ctx, owner, repo, filePath, &github.RepositoryContentGetOptions{Ref: branch})
	if err != nil {
		return "", fmt.Errorf("failed to get file contents from GitHub: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code from GitHub API: %d", resp.StatusCode)
	}

	if fileContent == nil {
		return "", fmt.Errorf("file not found at path: %s", filePath)
	}

	// GetSHA()でファイルのSHAを取得可能
	sha := fileContent.GetSHA()
	if strings.TrimSpace(sha) == "" {
		return "", fmt.Errorf("no SHA found for file: %s", filePath)
	}

	return sha, nil
}
