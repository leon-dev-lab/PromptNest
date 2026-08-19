<script setup>
import { onMounted } from "vue"
import { initPromptNestUI } from "./ui.js"

onMounted(() => {
  initPromptNestUI()
})
</script>

<template>
  <div class="promptnest-root">

  <div class="app-shell">
    <aside class="sidebar">
      <div class="brand">
        <div class="brand-mark"><span>PN</span></div>
        <div>
          <div class="brand-name">PromptNest</div>
          <div class="brand-sub">Prompt workspace</div>
        </div>
      </div>

      <button class="new-button" id="newPromptBtn" type="button">
        <span class="new-plus">+</span>
        <span>新建提示词</span>
        <kbd>Ctrl N</kbd>
      </button>

      <nav class="nav-block" aria-label="主导航">
        <button class="nav-item active" data-scope="all" type="button">
          <span class="nav-icon" data-icon="grid"></span>
          <span>全部提示词</span>
          <span class="nav-count" id="allCount">0</span>
        </button>
        <button class="nav-item" data-scope="favorites" type="button">
          <span class="nav-icon" data-icon="star"></span>
          <span>我的收藏</span>
          <span class="nav-count" id="favoriteCount">0</span>
        </button>
        <button class="nav-item" data-scope="recent" type="button">
          <span class="nav-icon" data-icon="clock"></span>
          <span>最近使用</span>
          <span class="nav-count" id="recentCount">0</span>
        </button>
      </nav>

      <div class="sidebar-heading">
        <span>分类</span>
        <button class="mini-manage" id="manageTaxonomyBtn" type="button" title="管理分类与标签">管理</button>
      </div>
      <div class="category-list" id="categoryList"></div>

      <div class="sidebar-spacer"></div>

      <div class="sidebar-tools">
        <button class="tool-row" id="importBtn" type="button"><span data-icon="download"></span><span>导入 JSON</span></button>
        <button class="tool-row" id="exportBtn" type="button"><span data-icon="upload"></span><span>导出备份</span></button>
        <button class="tool-row" id="folderBtn" type="button"><span data-icon="folder"></span><span>数据存储</span></button>
        <button class="tool-row github-row" id="githubBtn" type="button" title="打开 leon-dev-lab GitHub 主页"><span data-icon="github"></span><span>GitHub · leon-dev-lab</span></button>
      </div>
    </aside>

    <main class="main-stage">
      <header class="topbar">
        <div class="search-wrap">
          <span class="search-icon" data-icon="search"></span>
          <input id="searchInput" type="search" autocomplete="off" placeholder="搜索标题、内容、标签、备注…" />
          <span class="search-shortcut">Ctrl F</span>
        </div>
        <div class="top-actions">
          <button class="round-button" id="trayBtn" type="button" aria-label="隐藏到托盘" title="隐藏到托盘"><span data-icon="tray"></span></button>
          <button class="round-button" id="themeBtn" type="button" aria-label="切换主题"><span data-icon="moon"></span></button>
        </div>
      </header>

      <section class="dashboard-head">
        <div>
          <div class="eyebrow">PROMPT LIBRARY</div>
          <h1 id="pageTitle">全部提示词</h1>
          <p id="pageSubtitle">拖动卡片即可自由排序；分类和标签也支持独立管理。</p>
        </div>
        <div class="stat-ribbon">
          <div class="stat-cell"><strong id="statTotal">0</strong><span>总提示词</span></div>
          <div class="stat-cell"><strong id="statFavorite">0</strong><span>收藏</span></div>
          <div class="stat-cell"><strong id="statUses">0</strong><span>累计复制</span></div>
        </div>
      </section>

      <section class="toolbar-row">
        <div class="filter-chips" id="filterChips"></div>
        <div class="result-label"><span id="resultCount">0</span> 条结果 <span class="drag-hint" id="dragHint">· 可拖动排序</span></div>
      </section>

      <section class="prompt-grid" id="promptGrid" aria-live="polite"></section>

      <section class="empty-state hidden" id="emptyState">
        <div class="empty-orbit"><span>✦</span></div>
        <h2>这里还没有匹配的提示词</h2>
        <p>换个筛选条件，或者新建一条属于你的 Prompt。</p>
        <button class="primary-action" id="emptyCreateBtn" type="button">新建提示词</button>
      </section>
    </main>
  </div>

  <div class="drawer-backdrop" id="drawerBackdrop"></div>
  <aside class="editor-drawer" id="editorDrawer" aria-hidden="true">
    <div class="drawer-top">
      <div>
        <div class="drawer-kicker">PROMPT EDITOR</div>
        <div class="drawer-state" id="drawerState">编辑提示词</div>
      </div>
      <button class="round-button" id="closeDrawerBtn" type="button" aria-label="关闭"><span data-icon="x"></span></button>
    </div>

    <div class="editor-scroll">
      <div class="title-line">
        <input class="title-input" id="editTitle" maxlength="120" placeholder="提示词标题" />
        <button class="favorite-big" id="editFavoriteBtn" type="button" aria-label="收藏"><span data-icon="star"></span></button>
      </div>

      <div class="meta-grid">
        <div class="field compact">
          <span>分类</span>
          <div class="category-select" id="editCategorySelect">
            <button class="category-select-trigger" id="editCategoryTrigger" type="button" aria-haspopup="listbox" aria-expanded="false">
              <span class="category-select-value"><i class="category-select-dot" id="editCategoryDot"></i><span id="editCategoryText">选择分类</span></span>
              <span class="category-select-chevron" aria-hidden="true"></span>
            </button>
            <div class="category-select-menu" id="editCategoryMenu" role="listbox" aria-label="选择分类"></div>
            <input id="editCategory" type="hidden" />
          </div>
        </div>
        <div class="field compact">
          <span>标签</span>
          <div class="tag-token-box" id="tagInputShell">
            <div class="tag-token-list" id="editTagChips"></div>
            <input class="tag-token-input" id="editTagInput" maxlength="80" autocomplete="off" placeholder="输入标签后按 Enter" />
            <input id="editTags" type="hidden" />
            <div class="tag-suggestion-menu" id="editTagSuggestions" role="listbox" aria-label="标签联想"></div>
          </div>
        </div>
      </div>

      <label class="field content-field">
        <span class="field-header"><span>提示词正文</span><span id="charCount">0 字符</span></span>
        <textarea id="editContent" spellcheck="false" placeholder="在这里编写 Prompt…"></textarea>
      </label>

      <label class="field notes-field">
        <span>备注</span>
        <textarea id="editNotes" placeholder="记录使用场景、变量说明或注意事项…"></textarea>
      </label>

      <div class="editor-metrics">
        <div><span>创建</span><strong id="createdAt">—</strong></div>
        <div><span>更新</span><strong id="updatedAt">—</strong></div>
        <div><span>复制</span><strong id="useCount">0 次</strong></div>
      </div>
    </div>

    <div class="drawer-actions">
      <div class="secondary-actions">
        <button class="ghost-action danger" id="deleteBtn" type="button"><span data-icon="trash"></span>删除</button>
        <button class="ghost-action" id="duplicateBtn" type="button"><span data-icon="copy"></span>创建副本</button>
      </div>
      <div class="primary-actions">
        <span class="autosave-status" id="autoSaveStatus" data-mode="saved">自动保存已开启</span>
        <button class="copy-action" id="copyBtn" type="button"><span data-icon="clipboard"></span>复制标题 + 正文</button>
      </div>
    </div>
  </aside>

  <div class="modal-backdrop" id="taxonomyModal" aria-hidden="true">
    <section class="taxonomy-panel" role="dialog" aria-modal="true" aria-labelledby="taxonomyTitle">
      <header class="taxonomy-head">
        <div>
          <div class="drawer-kicker">LIBRARY ORGANIZER</div>
          <h2 id="taxonomyTitle">分类与标签</h2>
          <p>分类和标签均可重命名、删除、拖动排序。</p>
        </div>
        <button class="round-button" id="closeTaxonomyBtn" type="button" aria-label="关闭"><span data-icon="x"></span></button>
      </header>
      <div class="taxonomy-body">
        <section class="taxonomy-column">
          <div class="taxonomy-title-row"><strong>分类</strong><span id="categoryManageCount">0</span></div>
          <div class="taxonomy-add"><input id="newCategoryInput" maxlength="60" placeholder="新增分类…" /><button id="addCategoryBtn" type="button">添加</button></div>
          <div class="category-color-field">
            <span>新分类颜色</span>
            <div class="category-color-palette compact" id="newCategoryPalette" aria-label="新分类颜色"></div>
          </div>
          <div class="taxonomy-list" id="categoryManageList"></div>
        </section>
        <section class="taxonomy-column">
          <div class="taxonomy-title-row"><strong>标签</strong><span id="tagManageCount">0</span></div>
          <div class="taxonomy-add"><input id="newTagInput" maxlength="60" placeholder="新增标签…" /><button id="addTagBtn" type="button">添加</button></div>
          <div class="taxonomy-list" id="tagManageList"></div>
        </section>
      </div>
    </section>
  </div>

  <div class="modal-backdrop" id="dataModal" aria-hidden="true">
    <section class="data-panel" role="dialog" aria-modal="true" aria-labelledby="dataTitle">
      <header class="taxonomy-head data-head">
        <div>
          <div class="drawer-kicker">DATA STORAGE</div>
          <h2 id="dataTitle">数据目录</h2>
          <p>提示词库会保存在所选目录的 <strong>prompts.json</strong> 中。可粘贴路径，也可以直接选择文件夹。</p>
        </div>
        <button class="round-button" id="closeDataBtn" type="button" aria-label="关闭"><span data-icon="x"></span></button>
      </header>
      <div class="data-body">
        <label class="data-path-field">
          <span>当前数据目录</span>
          <input id="dataDirectoryInput" type="text" spellcheck="false" autocomplete="off" placeholder="选择或输入数据目录路径" />
        </label>
        <div class="data-path-note">
          <span id="dataDirectoryMode">默认目录</span>
          <code id="dataFilePath">—</code>
        </div>
        <div class="data-actions-row">
          <button class="data-action primary" id="applyDataPathBtn" type="button">应用路径</button>
          <button class="data-action" id="chooseDataPathBtn" type="button"><span data-icon="folder"></span>选择文件夹</button>
          <button class="data-action" id="openDataPathBtn" type="button">打开当前目录</button>
          <button class="data-action subtle" id="resetDataPathBtn" type="button">恢复默认</button>
        </div>
        <div class="data-help">
          <strong>切换规则</strong>
          <p>空目录会自动复制当前提示词库；如果目标目录已有 prompts.json，PromptNest 会先询问是否加载已有数据，不会静默覆盖。</p>
        </div>
      </div>
    </section>
  </div>

  <div class="modal-backdrop color-dialog-backdrop" id="colorDialog" aria-hidden="true">
    <section class="color-dialog-panel" role="dialog" aria-modal="true" aria-labelledby="colorDialogTitle">
      <header class="color-dialog-head">
        <div>
          <div class="drawer-kicker">CATEGORY COLOR</div>
          <h2 id="colorDialogTitle">选择分类颜色</h2>
          <p>正在设置 <strong id="colorDialogCategory">—</strong>。固定 20 种颜色，点击即可应用。</p>
        </div>
        <button class="round-button" id="closeColorDialogBtn" type="button" aria-label="关闭"><span data-icon="x"></span></button>
      </header>
      <div class="color-dialog-body">
        <div class="category-color-palette large" id="colorDialogPalette"></div>
        <p class="color-dialog-note">颜色会同步用于左侧分类圆点、顶部分类筛选、提示词卡片色条和分类徽标。</p>
      </div>
    </section>
  </div>

  <div class="modal-backdrop app-dialog-backdrop" id="appDialog" aria-hidden="true">
    <section class="app-dialog-panel" role="dialog" aria-modal="true" aria-labelledby="appDialogTitle" aria-describedby="appDialogMessage">
      <div class="app-dialog-icon" id="appDialogIcon" aria-hidden="true">!</div>
      <div class="app-dialog-copy">
        <h2 id="appDialogTitle">确认操作</h2>
        <p id="appDialogMessage">请确认是否继续。</p>
        <input class="app-dialog-input" id="appDialogInput" type="text" autocomplete="off" spellcheck="false" />
      </div>
      <div class="app-dialog-actions">
        <button class="app-dialog-button secondary" id="appDialogCancel" type="button">取消</button>
        <button class="app-dialog-button primary" id="appDialogConfirm" type="button">确定</button>
      </div>
    </section>
  </div>

  <div class="toast-stack" id="toastStack"></div>
  
  </div>
</template>
