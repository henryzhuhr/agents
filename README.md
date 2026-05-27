# skills

## 目录

- [快速开始](#快速开始)
- [技能列表](#技能列表)
  - [git-commit](#-git-commit)
  - [git-worktree](#-git-worktree)

## 快速开始

查看此仓库中所有可用的技能

```bash
npx skills add henryzhuhr/skills --list
```

通过命令安装到当前项目

```bash
npx skills add henryzhuhr/skills --copy --agent claude-code -y --skill <skill-name>
```

更新技能只能通过重新安装的方式更新，安装时会覆盖之前的技能文件，无法使用 `npx skills update` 的原因参考 PR： [#544](https://github.com/vercel-labs/skills/pull/544)、[#637](https://github.com/vercel-labs/skills/pull/637)、[#690](https://github.com/vercel-labs/skills/issues/690)，这是原项目 [vercel-labs/skills](https://github.com/vercel-labs/skills) 没有支持的功能。

## 技能列表

> 💡 安装技能的注意事项：
>
> - 不加 `-g` 参数将 `Installation scope` 设置为 `Project`，表示技能将安装在当前项目中
> - 使用 `--copy` 参数将 `Installation method` 设置为 `Copy`，表示技能将被复制到当前项目

- 使用 `--agent claude-code` 参数将自动设置安装目录为 `./claude/skills`

### 🧭 go-code-review-checklist

[`SKILL.md`](./skills/go-code-review-checklist/SKILL.md)：Go 代码审查 checklist，用于审查 Go 项目的变更、PR、diff 或提交，按正确性、错误处理、并发、context、接口设计、测试、安全、性能和可维护性系统发现问题。

```bash
npx skills add henryzhuhr/skills --copy --agent claude-code -y --skill go-code-review-checklist
```

### 🛠️ skill-creator

[`SKILL.md`](./skills/skill-creator/SKILL.md)：创建、修改和优化技能，支持运行 evals 评估技能性能，基准测试和描述优化。

```bash
npx skills add henryzhuhr/skills --copy --agent claude-code -y --skill skill-creator
```

### 📝 git-commit

[`SKILL.md`](./skills/git-commit/SKILL.md)：Use When 提交代码到 git 仓库，使用 commit message

```bash
npx skills add henryzhuhr/skills --copy --agent claude-code -y --skill git-commit
```

### 🌳 git-worktree

[`SKILL.md`](./skills/git-worktree/SKILL.md)：检查、创建和管理 Git worktree，用于并行分支开发。适用于需要按本地分支模式批量创建 worktree、把 worktree 放到当前仓库的 .worktrees 目录、避免为已在其他目录检出的分支重复创建 worktree、核对分支与 worktree 路径映射，或在多个分支上修改前先准备隔离工作区的场景。

```bash
npx skills add henryzhuhr/skills --copy --agent claude-code -y --skill git-worktree
```
