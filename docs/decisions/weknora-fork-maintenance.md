# WeKnora Fork 维护规则

- 状态：已接受
- 日期：2026-10-20
- 负责人：王
- Fork：`macheal/WeKnora`
- 上游：`Tencent/WeKnora`

## 决策

WeKnora 作为独立 fork 长期维护。生产环境中的扩展基于实际部署到 Kubernetes 的发布版本，当前为 `v0.8.3`。面向上游的贡献在临时分支上准备，该分支基于最新的 `upstream/main`，不直接在生产扩展分支上开发上游 PR。

## 分支模型

| 分支 | 基线 | 用途 | 合并规则 |
| --- | --- | --- | --- |
| `main` | `upstream/main` | 保持与 Tencent 一致的干净镜像 | 仅允许快进合并 |
| `codex/weknora-extensions` | 发布 tag，当前为 `v0.8.3` | 长期维护的生产扩展 | 合并功能分支，不做 rebase |
| `feat/*`、`fix/*` | `codex/weknora-extensions` | 内部功能与修复 | 合并回扩展分支 |
| `upstream/feat/*`、`upstream/fix/*` | `upstream/main` | 提交给上游的最小 PR 改动 | rebase 后提交给 Tencent |

生产扩展分支与上游 PR 分支必须分开。这样在上游版本完成升级验证之前，`upstream/main` 的新改动不会进入生产扩展线。

## 版本与升级规则

兼容性基线是已经部署的发布版本，而不是尚未打 tag 的最新 `main` 提交。

1. 先部署并验证上游新版本，再切换扩展基线。
2. 在临时分支上升级：
   `git switch -c chore/upgrade-vX.Y.Z codex/weknora-extensions`。
3. 将新的上游发布 tag 或 main 快照合并到该分支。
4. 在升级分支中解决冲突，并运行对应版本的迁移与回归检查。
5. 将升级分支合并回 `codex/weknora-extensions`。
6. 创建下游发布 tag，例如 `v0.8.3-macheal.1`；不得覆盖上游 tag。

`main` 必须保持为干净的上游镜像，不得提交本地扩展。

## 上游 PR 流程

向 `Tencent/WeKnora:main` 提交 PR 前，执行：

```bash
git fetch upstream --prune --tags
git switch -c upstream/feat/feature-name upstream/main
# 只放入适合提交上游的提交
git merge upstream/main
# 在本地解决冲突并运行针对性测试
git push -u origin upstream/feat/feature-name
```

PR 的目标分支固定为 `Tencent/WeKnora:main`。PR 不得包含内部地址、凭证、产品专用策略或与本次改动无关的下游内容。

上游 PR 合并后，再通过独立的迁移改动将结果同步到扩展线。不能因为上游已经合并，就默认下游扩展分支已经可以进入生产。

## 兼容性隔离

优先采用容易跨版本维护的改动方式：

- 新能力优先放入新文件、新包、新路由或适配层。
- 对核心文件的改动保持小而集中，并记录所涉及的上游文件与提交。
- 未完成的行为放到默认关闭的功能开关之后。
- 数据库迁移保持增量式，并同时验证 PostgreSQL 与 SQLite 路径。
- 在升级 PR 中记录冲突、迁移步骤和验证结果。

## 必须验证的内容

扩展改动需要运行受影响包的测试；涉及 UI 时还要运行前端构建。版本升级还需要额外运行迁移检查、记忆相关回归测试和部署配置冒烟测试。环境相关失败必须记录，不能省略。

## 当前基线

- 生产发布版本：`v0.8.3`
- 发布提交：`3e22822c1d2397be730c458c7539d0eb8c113a24`
- 生产扩展分支：`codex/weknora-extensions`
- 部署目标：本地 Kubernetes
