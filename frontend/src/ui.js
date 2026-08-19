export function initPromptNestUI() {
  // PromptNest dialogs must always use the in-app modal. If legacy code accidentally
  // invokes a browser-native dialog, suppress it so "wails.localhost 显示" never leaks into the UI.
  window.alert = (message) => { console.warn("Blocked native alert:", message); };
  window.confirm = (message) => { console.warn("Blocked native confirm:", message); return false; };
  window.prompt = (message) => { console.warn("Blocked native prompt:", message); return null; };

  "use strict";

  const COMMON_CATEGORY = "通用分类";
  const AUTO_SAVE_DELAY = 650;
  const DEFAULT_CATEGORY_PALETTE = [
    "#D96C4B", "#6F8B5B", "#7F6FB2", "#4D8C9C", "#C29342",
    "#5F78B5", "#B76479", "#4F9A7B", "#D9895B", "#839750",
    "#5D91B8", "#B77B45", "#8B7665", "#6F9AA4", "#A66C9B",
    "#67A18D", "#E08A72", "#9A8B52", "#6A87A9", "#AA745C"
  ];
  let autoSaveTimer = null;
  let autoSaveDirty = false;
  let tagSuggestionIndex = -1;

  const state = {
    prompts: [],
    categories: [],
    categoryColors: {},
    categoryPalette: [...DEFAULT_CATEGORY_PALETTE],
    newCategoryColor: DEFAULT_CATEGORY_PALETTE[0],
    colorEditingCategory: null,
    tags: [],
    isDarkTheme: false,
    scope: "all",
    category: "全部分类",
    search: "",
    editingId: null,
    dataPath: "",
    dataDirectory: "",
    defaultDataDirectory: "",
    customDataDirectory: false,
    draggedPromptId: null
  };

  const $ = (id) => document.getElementById(id);
  const els = {
    newPromptBtn: $("newPromptBtn"), emptyCreateBtn: $("emptyCreateBtn"), searchInput: $("searchInput"), themeBtn: $("themeBtn"), trayBtn: $("trayBtn"),
    allCount: $("allCount"), favoriteCount: $("favoriteCount"), recentCount: $("recentCount"), categoryList: $("categoryList"),
    filterChips: $("filterChips"), promptGrid: $("promptGrid"), emptyState: $("emptyState"), resultCount: $("resultCount"), dragHint: $("dragHint"),
    pageTitle: $("pageTitle"), pageSubtitle: $("pageSubtitle"), statTotal: $("statTotal"), statFavorite: $("statFavorite"), statUses: $("statUses"),
    importBtn: $("importBtn"), exportBtn: $("exportBtn"), folderBtn: $("folderBtn"), githubBtn: $("githubBtn"), manageTaxonomyBtn: $("manageTaxonomyBtn"),
    editorDrawer: $("editorDrawer"), drawerBackdrop: $("drawerBackdrop"), closeDrawerBtn: $("closeDrawerBtn"), drawerState: $("drawerState"),
    editTitle: $("editTitle"), editCategory: $("editCategory"), editCategorySelect: $("editCategorySelect"), editCategoryTrigger: $("editCategoryTrigger"), editCategoryMenu: $("editCategoryMenu"), editCategoryText: $("editCategoryText"), editCategoryDot: $("editCategoryDot"),
    editTags: $("editTags"), editTagInput: $("editTagInput"), editTagChips: $("editTagChips"), tagInputShell: $("tagInputShell"), editTagSuggestions: $("editTagSuggestions"), editContent: $("editContent"), editNotes: $("editNotes"),
    editFavoriteBtn: $("editFavoriteBtn"), charCount: $("charCount"), createdAt: $("createdAt"), updatedAt: $("updatedAt"), useCount: $("useCount"),
    deleteBtn: $("deleteBtn"), duplicateBtn: $("duplicateBtn"), copyBtn: $("copyBtn"), autoSaveStatus: $("autoSaveStatus"), toastStack: $("toastStack"),
    taxonomyModal: $("taxonomyModal"), closeTaxonomyBtn: $("closeTaxonomyBtn"), categoryManageList: $("categoryManageList"), tagManageList: $("tagManageList"),
    categoryManageCount: $("categoryManageCount"), tagManageCount: $("tagManageCount"), newCategoryInput: $("newCategoryInput"), addCategoryBtn: $("addCategoryBtn"), newCategoryPalette: $("newCategoryPalette"),
    newTagInput: $("newTagInput"), addTagBtn: $("addTagBtn"),
    colorDialog: $("colorDialog"), colorDialogCategory: $("colorDialogCategory"), colorDialogPalette: $("colorDialogPalette"), closeColorDialogBtn: $("closeColorDialogBtn"),
    dataModal: $("dataModal"), closeDataBtn: $("closeDataBtn"), dataDirectoryInput: $("dataDirectoryInput"), dataDirectoryMode: $("dataDirectoryMode"),
    dataFilePath: $("dataFilePath"), applyDataPathBtn: $("applyDataPathBtn"), chooseDataPathBtn: $("chooseDataPathBtn"), openDataPathBtn: $("openDataPathBtn"), resetDataPathBtn: $("resetDataPathBtn"),
    appDialog: $("appDialog"), appDialogPanel: document.querySelector("#appDialog .app-dialog-panel"), appDialogIcon: $("appDialogIcon"),
    appDialogTitle: $("appDialogTitle"), appDialogMessage: $("appDialogMessage"), appDialogInput: $("appDialogInput"),
    appDialogCancel: $("appDialogCancel"), appDialogConfirm: $("appDialogConfirm")
  };

  const icons = {
    grid: '<svg viewBox="0 0 24 24"><rect x="3" y="3" width="7" height="7" rx="2"/><rect x="14" y="3" width="7" height="7" rx="2"/><rect x="3" y="14" width="7" height="7" rx="2"/><rect x="14" y="14" width="7" height="7" rx="2"/></svg>',
    star: '<svg viewBox="0 0 24 24"><path d="m12 3 2.8 5.7 6.2.9-4.5 4.4 1.1 6.2-5.6-2.9-5.6 2.9 1.1-6.2L3 9.6l6.2-.9L12 3Z"/></svg>',
    clock: '<svg viewBox="0 0 24 24"><circle cx="12" cy="12" r="9"/><path d="M12 7v5l3 2"/></svg>',
    search: '<svg viewBox="0 0 24 24"><circle cx="11" cy="11" r="7"/><path d="m20 20-4-4"/></svg>',
    moon: '<svg viewBox="0 0 24 24"><path d="M20 15.5A8.5 8.5 0 0 1 8.5 4 8.5 8.5 0 1 0 20 15.5Z"/></svg>',
    sun: '<svg viewBox="0 0 24 24"><circle cx="12" cy="12" r="4"/><path d="M12 2v2M12 20v2M4.9 4.9l1.4 1.4M17.7 17.7l1.4 1.4M2 12h2M20 12h2M4.9 19.1l1.4-1.4M17.7 6.3l1.4-1.4"/></svg>',
    x: '<svg viewBox="0 0 24 24"><path d="m6 6 12 12M18 6 6 18"/></svg>',
    trash: '<svg viewBox="0 0 24 24"><path d="M4 7h16M9 7V4h6v3M7 7l1 13h8l1-13M10 11v5M14 11v5"/></svg>',
    copy: '<svg viewBox="0 0 24 24"><rect x="8" y="8" width="11" height="11" rx="2"/><path d="M16 8V6a2 2 0 0 0-2-2H6a2 2 0 0 0-2 2v8a2 2 0 0 0 2 2h2"/></svg>',
    clipboard: '<svg viewBox="0 0 24 24"><path d="M9 5H6a2 2 0 0 0-2 2v12a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V7a2 2 0 0 0-2-2h-3"/><rect x="9" y="3" width="6" height="4" rx="1"/></svg>',
    folder: '<svg viewBox="0 0 24 24"><path d="M3 6.5A2.5 2.5 0 0 1 5.5 4H10l2 2h6.5A2.5 2.5 0 0 1 21 8.5v8A2.5 2.5 0 0 1 18.5 19h-13A2.5 2.5 0 0 1 3 16.5v-10Z"/></svg>',
    upload: '<svg viewBox="0 0 24 24"><path d="M12 16V4M7.5 8.5 12 4l4.5 4.5M5 20h14"/></svg>',
    download: '<svg viewBox="0 0 24 24"><path d="M12 4v12M7.5 11.5 12 16l4.5-4.5M5 20h14"/></svg>',
    tray: '<svg viewBox="0 0 24 24"><path d="M4 14h4l2 3h4l2-3h4v5H4v-5Z"/><path d="M12 4v9M8.5 9.5 12 13l3.5-3.5"/></svg>',
    edit: '<svg viewBox="0 0 24 24"><path d="M4 20h4l11-11-4-4L4 16v4Z"/><path d="m13.5 6.5 4 4"/></svg>',
    github: '<svg viewBox="0 0 24 24"><path d="M12 2.7a9.5 9.5 0 0 0-3 18.5c.5.1.7-.2.7-.5v-1.9c-2.8.6-3.4-1.2-3.4-1.2-.5-1.2-1.1-1.5-1.1-1.5-.9-.6.1-.6.1-.6 1 0 1.6 1.1 1.6 1.1.9 1.6 2.4 1.1 3 .9.1-.7.4-1.1.7-1.3-2.2-.3-4.6-1.1-4.6-4.8 0-1.1.4-1.9 1-2.6-.1-.3-.4-1.3.1-2.6 0 0 .8-.3 2.7 1a9.3 9.3 0 0 1 4.9 0c1.9-1.3 2.7-1 2.7-1 .5 1.3.2 2.3.1 2.6.6.7 1 1.6 1 2.6 0 3.7-2.3 4.5-4.6 4.8.4.3.7.9.7 1.7v2.8c0 .3.2.6.7.5A9.5 9.5 0 0 0 12 2.7Z"/></svg>'
  };

  document.querySelectorAll("[data-icon]").forEach(el => { el.innerHTML = icons[el.dataset.icon] || ""; });

  async function post(type, payload = {}) {
    const handle = window.go?.main?.App?.HandleMessage;
    if (!handle) return;
    try {
      const raw = await handle(JSON.stringify({ type, payload }));
      if (!raw) return;
      const messages = JSON.parse(raw);
      (Array.isArray(messages) ? messages : [messages]).forEach(handleHostMessage);
    } catch (error) {
      toast(error?.message || String(error), "error");
    }
  }

  function safe(value) {
    return String(value ?? "").replace(/[&<>'"]/g, c => ({ "&": "&amp;", "<": "&lt;", ">": "&gt;", "'": "&#39;", '"': "&quot;" }[c]));
  }

  function normalizePrompt(p, index = 0) {
    return {
      id: String(p.id || p.Id || ""),
      title: p.title ?? p.Title ?? "未命名提示词",
      category: p.category ?? p.Category ?? COMMON_CATEGORY,
      tags: p.tags ?? p.Tags ?? "",
      content: p.content ?? p.Content ?? "",
      notes: p.notes ?? p.Notes ?? "",
      isFavorite: p.isFavorite ?? p.IsFavorite ?? false,
      sortOrder: Number(p.sortOrder ?? p.SortOrder ?? index),
      createdAt: p.createdAt ?? p.CreatedAt,
      updatedAt: p.updatedAt ?? p.UpdatedAt,
      lastUsedAt: p.lastUsedAt ?? p.LastUsedAt,
      useCount: p.useCount ?? p.UseCount ?? 0
    };
  }

  function distinct(values) {
    const seen = new Set();
    const result = [];
    values.forEach(v => {
      const value = String(v || "").trim();
      const key = value.toLocaleLowerCase();
      if (value && !seen.has(key)) { seen.add(key); result.push(value); }
    });
    return result;
  }

  function categoryOrder(values) {
    return distinct(values);
  }

  function splitTags(value) {
    return String(value || "").split(/[,，]/).map(x => x.trim()).filter(Boolean);
  }

  function categoryNames() {
    const names = distinct([...state.categories, ...state.prompts.map(p => p.category || COMMON_CATEGORY)]);
    return names;
  }

  function tagNames() {
    // The backend keeps this list authoritative and removes a deleted tag from
    // every prompt. Do not rebuild it from stale editor/prompt snapshots here.
    return distinct(state.tags);
  }

  function categoriesWithCounts() {
    const counts = new Map();
    state.prompts.forEach(p => {
      const name = (p.category || COMMON_CATEGORY).trim() || COMMON_CATEGORY;
      counts.set(name.toLocaleLowerCase(), (counts.get(name.toLocaleLowerCase()) || 0) + 1);
    });
    return categoryNames().map(name => [name, counts.get(name.toLocaleLowerCase()) || 0]);
  }

  function filteredPrompts() {
    const needle = state.search.trim().toLocaleLowerCase();
    return state.prompts
      .filter(p => {
        if (state.scope === "favorites" && !p.isFavorite) return false;
        if (state.scope === "recent" && !p.lastUsedAt) return false;
        if (state.category !== "全部分类" && p.category !== state.category) return false;
        if (!needle) return true;
        return [p.title, p.category, p.tags, p.content, p.notes].some(v => String(v || "").toLocaleLowerCase().includes(needle));
      })
      .sort((a, b) => {
        if (state.scope === "recent") return new Date(b.lastUsedAt || 0) - new Date(a.lastUsedAt || 0);
        return Number(a.sortOrder || 0) - Number(b.sortOrder || 0);
      });
  }

  function applyTheme() {
    document.documentElement.dataset.theme = state.isDarkTheme ? "dark" : "light";
    els.themeBtn.querySelector("span").innerHTML = state.isDarkTheme ? icons.sun : icons.moon;
    els.themeBtn.title = state.isDarkTheme ? "切换浅色模式" : "切换深色模式";
  }

  function formatDate(value) {
    if (!value) return "—";
    const d = new Date(value);
    if (Number.isNaN(d.getTime())) return "—";
    return new Intl.DateTimeFormat("zh-CN", { month: "2-digit", day: "2-digit", hour: "2-digit", minute: "2-digit" }).format(d);
  }

  function relativeDate(value) {
    if (!value) return "从未使用";
    const d = new Date(value);
    const diff = Date.now() - d.getTime();
    const day = Math.floor(diff / 86400000);
    if (day <= 0) return "今天";
    if (day === 1) return "昨天";
    if (day < 30) return `${day} 天前`;
    return formatDate(value).split(" ")[0];
  }

  function paletteColors() {
    return (Array.isArray(state.categoryPalette) && state.categoryPalette.length ? state.categoryPalette : DEFAULT_CATEGORY_PALETTE).slice(0, 20);
  }

  function categoryColor(category) {
    const name = String(category || COMMON_CATEGORY).trim() || COMMON_CATEGORY;
    const key = Object.keys(state.categoryColors || {}).find(k => k.toLocaleLowerCase() === name.toLocaleLowerCase());
    if (key && state.categoryColors[key]) return state.categoryColors[key];
    const names = categoryNames();
    const index = Math.max(0, names.findIndex(x => x.toLocaleLowerCase() === name.toLocaleLowerCase()));
    const palette = paletteColors();
    return palette[index % palette.length] || "#6F8B5B";
  }

  function nextUnusedCategoryColor() {
    const used = new Set(categoryNames().map(name => categoryColor(name).toLocaleLowerCase()));
    const palette = paletteColors();
    return palette.find(color => !used.has(color.toLocaleLowerCase())) || palette[categoryNames().length % palette.length] || "#6F8B5B";
  }

  function colorSwatchesHtml(selected, attribute = "data-new-category-color") {
    return paletteColors().map((color, index) => `<button type="button" class="category-color-swatch ${String(selected).toLocaleLowerCase() === color.toLocaleLowerCase() ? "selected" : ""}" ${attribute}="${color}" style="--swatch:${color}" title="颜色 ${index + 1}" aria-label="选择颜色 ${index + 1}"></button>`).join("");
  }

  function renderNewCategoryPalette() {
    if (!els.newCategoryPalette) return;
    if (!paletteColors().some(c => c.toLocaleLowerCase() === String(state.newCategoryColor).toLocaleLowerCase())) state.newCategoryColor = nextUnusedCategoryColor();
    els.newCategoryPalette.innerHTML = colorSwatchesHtml(state.newCategoryColor);
    els.newCategoryPalette.querySelectorAll("[data-new-category-color]").forEach(btn => btn.addEventListener("click", () => {
      state.newCategoryColor = btn.dataset.newCategoryColor;
      renderNewCategoryPalette();
    }));
  }

  function openColorDialog(category) {
    if (!category) return;
    state.colorEditingCategory = category;
    els.colorDialogCategory.textContent = category;
    const selected = categoryColor(category);
    els.colorDialogPalette.innerHTML = colorSwatchesHtml(selected, "data-category-color-choice");
    els.colorDialogPalette.querySelectorAll("[data-category-color-choice]").forEach(btn => btn.addEventListener("click", () => {
      post("setCategoryColor", { name: category, color: btn.dataset.categoryColorChoice });
      closeColorDialog();
    }));
    els.colorDialog.classList.add("open");
    els.colorDialog.setAttribute("aria-hidden", "false");
  }

  function closeColorDialog() {
    state.colorEditingCategory = null;
    els.colorDialog.classList.remove("open");
    els.colorDialog.setAttribute("aria-hidden", "true");
  }

  function cardAccent(category) { return categoryColor(category); }

  function render() {
    applyTheme();
    renderSidebar();
    renderHeader();
    renderGrid();
    if (state.editingId) fillEditor(false);
    if (els.taxonomyModal.classList.contains("open")) renderTaxonomy();
    if (els.dataModal.classList.contains("open")) renderDataSettings();
  }

  function renderSidebar() {
    els.allCount.textContent = state.prompts.length;
    els.favoriteCount.textContent = state.prompts.filter(p => p.isFavorite).length;
    els.recentCount.textContent = state.prompts.filter(p => p.lastUsedAt).length;

    document.querySelectorAll(".nav-item[data-scope]").forEach(btn => btn.classList.toggle("active", btn.dataset.scope === state.scope));

    els.categoryList.innerHTML = categoriesWithCounts().map(([name, count]) => `
      <button type="button" class="category-item ${state.category === name ? "active" : ""}" data-category="${encodeURIComponent(name)}" style="--category-color:${categoryColor(name)}">
        <span>${safe(name)}</span><span>${count}</span>
      </button>`).join("");

    els.categoryList.querySelectorAll(".category-item").forEach(btn => btn.addEventListener("click", () => {
      state.scope = "all";
      state.category = decodeURIComponent(btn.dataset.category || "") || "全部分类";
      render();
    }));
  }

  function renderHeader() {
    const favs = state.prompts.filter(p => p.isFavorite).length;
    const uses = state.prompts.reduce((n, p) => n + Number(p.useCount || 0), 0);
    els.statTotal.textContent = state.prompts.length;
    els.statFavorite.textContent = favs;
    els.statUses.textContent = uses;

    let title = "全部提示词";
    let subtitle = "拖动卡片即可自由排序；分类和标签也支持独立管理。";
    if (state.scope === "favorites") { title = "我的收藏"; subtitle = "收藏视图同样可以拖动，调整它们在总库中的相对顺序。"; }
    else if (state.scope === "recent") { title = "最近使用"; subtitle = "最近使用按时间排序，因此此视图不启用手动拖拽。"; }
    else if (state.category !== "全部分类") { title = state.category; subtitle = `正在查看“${state.category}”分类；拖动可调整该分类提示词的相对位置。`; }
    els.pageTitle.textContent = title;
    els.pageSubtitle.textContent = subtitle;
    els.dragHint.classList.toggle("hidden", state.scope === "recent");

    const cats = categoryNames().slice(0, 8);
    els.filterChips.innerHTML = `<button class="filter-chip ${state.category === "全部分类" ? "active" : ""}" data-chip="${encodeURIComponent("全部分类")}">全部</button>` +
      cats.map(c => `<button class="filter-chip category-filter ${state.category === c ? "active" : ""}" data-chip="${encodeURIComponent(c)}" style="--category-color:${categoryColor(c)}">${safe(c)}</button>`).join("");
    els.filterChips.querySelectorAll(".filter-chip").forEach(btn => btn.addEventListener("click", () => {
      state.scope = "all";
      state.category = decodeURIComponent(btn.dataset.chip || "") || "全部分类";
      render();
    }));
  }

  function renderGrid() {
    const list = filteredPrompts();
    const canDrag = state.scope !== "recent";
    els.resultCount.textContent = list.length;
    els.promptGrid.classList.toggle("hidden", list.length === 0);
    els.emptyState.classList.toggle("hidden", list.length !== 0);

    els.promptGrid.innerHTML = list.map(p => {
      const tags = splitTags(p.tags).slice(0, 3);
      const preview = (p.content || "还没有内容").replace(/\s+/g, " ").trim();
      return `<article class="prompt-card" data-id="${p.id}" draggable="${canDrag ? "true" : "false"}" style="--card-accent:${cardAccent(p.category)};--category-color:${categoryColor(p.category)}">
        ${canDrag ? '<span class="drag-grip" aria-hidden="true"></span>' : ""}
        <div class="card-top">
          <span class="category-badge">${safe(p.category || COMMON_CATEGORY)}</span>
          <button class="card-fav ${p.isFavorite ? "active" : ""}" type="button" data-fav="${p.id}" aria-label="收藏">${icons.star}</button>
        </div>
        <h3 class="card-title">${safe(p.title)}</h3>
        <p class="card-preview">${safe(preview)}</p>
        <div class="card-spacer"></div>
        <div class="card-tags">${tags.map(t => `<span class="tag">${safe(t)}</span>`).join("")}</div>
        <div class="card-footer">
          <span>${p.lastUsedAt ? `最近 ${relativeDate(p.lastUsedAt)}` : `更新 ${relativeDate(p.updatedAt)}`}</span>
          <span class="card-uses"><span>${Number(p.useCount || 0)} 次</span><button type="button" class="quick-copy" data-copy="${p.id}">复制</button></span>
        </div>
      </article>`;
    }).join("");

    els.promptGrid.querySelectorAll(".prompt-card").forEach(card => {
      card.addEventListener("click", e => {
        if (e.target.closest("button") || card.classList.contains("dragging")) return;
        openEditor(card.dataset.id);
      });
      if (canDrag) wirePromptDrag(card);
    });
    els.promptGrid.querySelectorAll("[data-fav]").forEach(btn => btn.addEventListener("click", e => {
      e.stopPropagation();
      post("toggleFavorite", { id: btn.dataset.fav });
    }));
    els.promptGrid.querySelectorAll("[data-copy]").forEach(btn => btn.addEventListener("click", e => {
      e.stopPropagation();
      post("copyPrompt", { id: btn.dataset.copy });
      toast("标题和正文已复制到剪贴板", "success");
    }));
  }

  function wirePromptDrag(card) {
    card.addEventListener("dragstart", e => {
      state.draggedPromptId = card.dataset.id;
      card.classList.add("dragging");
      e.dataTransfer.effectAllowed = "move";
      e.dataTransfer.setData("text/plain", card.dataset.id || "");
    });
    card.addEventListener("dragover", e => {
      if (!state.draggedPromptId || state.draggedPromptId === card.dataset.id) return;
      e.preventDefault();
      e.dataTransfer.dropEffect = "move";
      card.classList.add("drag-over");
    });
    card.addEventListener("dragleave", () => card.classList.remove("drag-over"));
    card.addEventListener("drop", e => {
      e.preventDefault();
      card.classList.remove("drag-over");
      reorderVisiblePrompts(state.draggedPromptId, card.dataset.id, e);
    });
    card.addEventListener("dragend", () => {
      card.classList.remove("dragging");
      els.promptGrid.querySelectorAll(".drag-over").forEach(x => x.classList.remove("drag-over"));
      state.draggedPromptId = null;
    });
  }

  function reorderVisiblePrompts(sourceId, targetId, event) {
    if (!sourceId || !targetId || sourceId === targetId) return;

    const visibleOrder = filteredPrompts().map(p => p.id);
    const sourceIndex = visibleOrder.indexOf(sourceId);
    const targetIndex = visibleOrder.indexOf(targetId);
    if (sourceIndex < 0 || targetIndex < 0) return;

    visibleOrder.splice(sourceIndex, 1);
    let insertAt = visibleOrder.indexOf(targetId);
    const targetCard = els.promptGrid.querySelector(`.prompt-card[data-id="${CSS.escape(targetId)}"]`);
    if (targetCard) {
      const rect = targetCard.getBoundingClientRect();
      const verticalAfter = event.clientY > rect.top + rect.height / 2;
      const horizontalAfter = Math.abs(event.clientY - (rect.top + rect.height / 2)) < rect.height * .24 &&
        event.clientX > rect.left + rect.width / 2;
      if (verticalAfter || horizontalAfter) insertAt += 1;
    }
    visibleOrder.splice(Math.max(0, insertAt), 0, sourceId);

    // Only replace the slots occupied by the currently visible prompts. This keeps
    // hidden/search-filtered prompts stable while still allowing local reordering.
    const globalOrder = [...state.prompts].sort((a, b) => Number(a.sortOrder) - Number(b.sortOrder));
    const visibleSet = new Set(visibleOrder);
    const visibleSlots = [];
    globalOrder.forEach((p, index) => { if (visibleSet.has(p.id)) visibleSlots.push(index); });
    const byId = new Map(state.prompts.map(p => [p.id, p]));
    visibleSlots.forEach((slot, index) => { globalOrder[slot] = byId.get(visibleOrder[index]); });
    globalOrder.forEach((p, index) => { p.sortOrder = index; });

    state.prompts = globalOrder;
    renderGrid();
    post("reorderPrompts", { ids: globalOrder.map(p => p.id) });
  }

  function closeEditorCategoryMenu() {
    els.editCategorySelect?.classList.remove("open");
    els.editCategoryTrigger?.setAttribute("aria-expanded", "false");
  }

  function setEditorCategory(name, save = false) {
    const names = categoryNames();
    const current = String(name || names[0] || COMMON_CATEGORY).trim() || COMMON_CATEGORY;
    els.editCategory.value = current;
    els.editCategoryText.textContent = current;
    els.editCategoryDot.style.setProperty("--category-color", categoryColor(current));
    els.editCategoryMenu.querySelectorAll("[data-editor-category]").forEach(btn => {
      btn.classList.toggle("selected", decodeURIComponent(btn.dataset.editorCategory || "") === current);
      btn.setAttribute("aria-selected", decodeURIComponent(btn.dataset.editorCategory || "") === current ? "true" : "false");
    });
    if (save) scheduleAutoSave();
  }

  function renderEditorCategoryOptions(selected) {
    const names = categoryNames();
    const current = String(selected || names[0] || COMMON_CATEGORY).trim() || COMMON_CATEGORY;
    if (!names.includes(current)) names.push(current);
    els.editCategoryMenu.innerHTML = names.map(name => `
      <button type="button" class="category-select-option" role="option" data-editor-category="${encodeURIComponent(name)}" style="--category-color:${categoryColor(name)}">
        <span class="category-select-option-dot"></span><span>${safe(name)}</span><span class="category-select-check">✓</span>
      </button>`).join("");
    els.editCategoryMenu.querySelectorAll("[data-editor-category]").forEach(btn => btn.addEventListener("click", () => {
      const name = decodeURIComponent(btn.dataset.editorCategory || "");
      setEditorCategory(name, true);
      closeEditorCategoryMenu();
      els.editCategoryTrigger.focus();
    }));
    setEditorCategory(current, false);
  }

  function editorTagValues() {
    return splitTags(els.editTags.value || "");
  }

  function renderEditorTags(value) {
    const tags = splitTags(value || "");
    els.editTags.value = tags.join(", ");
    els.editTagChips.innerHTML = tags.map(tag => `
      <span class="tag-token" data-editor-tag="${encodeURIComponent(tag)}">
        <span>${safe(tag)}</span><button type="button" aria-label="删除标签 ${safe(tag)}" title="删除标签">×</button>
      </span>`).join("");
    els.editTagChips.querySelectorAll("[data-editor-tag] button").forEach(btn => btn.addEventListener("click", () => {
      const chip = btn.closest("[data-editor-tag]");
      const removed = decodeURIComponent(chip?.dataset.editorTag || "");
      const next = editorTagValues().filter(tag => tag.toLocaleLowerCase() !== removed.toLocaleLowerCase());
      renderEditorTags(next.join(", "));
      scheduleAutoSave();
      els.editTagInput.focus();
      renderTagSuggestions();
    }));
  }

  function tagUsageCount(name) {
    const key = String(name || "").toLocaleLowerCase();
    return state.prompts.reduce((count, prompt) => count + (splitTags(prompt.tags).some(tag => tag.toLocaleLowerCase() === key) ? 1 : 0), 0);
  }

  function getTagSuggestions() {
    const query = String(els.editTagInput.value || "").trim().toLocaleLowerCase();
    if (!query) return [];
    const selected = new Set(editorTagValues().map(tag => tag.toLocaleLowerCase()));
    return tagNames()
      .filter(name => !selected.has(name.toLocaleLowerCase()))
      .map((name, order) => {
        const lower = name.toLocaleLowerCase();
        const starts = lower.startsWith(query);
        const at = lower.indexOf(query);
        return { name, order, starts, at };
      })
      .filter(item => item.at >= 0)
      .sort((a, b) => Number(b.starts) - Number(a.starts) || a.at - b.at || a.order - b.order)
      .slice(0, 8)
      .map(item => item.name);
  }

  function closeTagSuggestions() {
    tagSuggestionIndex = -1;
    els.editTagSuggestions.classList.remove("open");
    els.editTagSuggestions.innerHTML = "";
  }

  function renderTagSuggestions() {
    const items = getTagSuggestions();
    if (!items.length || document.activeElement !== els.editTagInput) { closeTagSuggestions(); return; }
    if (tagSuggestionIndex >= items.length) tagSuggestionIndex = items.length - 1;
    els.editTagSuggestions.innerHTML = items.map((name, index) => `
      <button type="button" class="tag-suggestion-item${index === tagSuggestionIndex ? " active" : ""}" data-tag-suggestion="${encodeURIComponent(name)}" role="option" aria-selected="${index === tagSuggestionIndex ? "true" : "false"}">
        <span class="tag-suggestion-name">${safe(name)}</span><span class="tag-suggestion-hint">${tagUsageCount(name)} 条</span>
      </button>`).join("");
    els.editTagSuggestions.classList.add("open");
    els.editTagSuggestions.querySelectorAll("[data-tag-suggestion]").forEach(btn => {
      btn.addEventListener("pointerdown", e => e.preventDefault());
      btn.addEventListener("click", () => selectTagSuggestion(decodeURIComponent(btn.dataset.tagSuggestion || "")));
    });
  }

  function selectTagSuggestion(name) {
    const value = String(name || "").trim();
    if (!value) return;
    const tags = editorTagValues();
    if (!tags.some(tag => tag.toLocaleLowerCase() === value.toLocaleLowerCase())) {
      tags.push(value);
      renderEditorTags(tags.join(", "));
      scheduleAutoSave();
    }
    els.editTagInput.value = "";
    closeTagSuggestions();
    els.editTagInput.focus();
  }

  function commitEditorTag() {
    let name = String(els.editTagInput.value || "").trim();
    // Storage is comma-separated internally, so commas typed into a single chip
    // are normalised to spaces. The user only needs Enter to commit a tag.
    name = name.replace(/[,，]+/g, " ").replace(/\s+/g, " ").trim();
    if (!name) { els.editTagInput.value = ""; return false; }
    const canonical = tagNames().find(tag => tag.toLocaleLowerCase() === name.toLocaleLowerCase());
    if (canonical) name = canonical;
    const tags = editorTagValues();
    if (tags.some(tag => tag.toLocaleLowerCase() === name.toLocaleLowerCase())) {
      els.editTagInput.value = "";
      toast("该标签已经存在");
      return false;
    }
    tags.push(name);
    els.editTagInput.value = "";
    renderEditorTags(tags.join(", "));
    closeTagSuggestions();
    scheduleAutoSave();
    return true;
  }

  function openEditor(id) {
    const nextId = String(id || "");
    if (state.editingId && state.editingId !== nextId) flushAutoSave();
    state.editingId = nextId;
    autoSaveDirty = false;
    clearTimeout(autoSaveTimer);
    setAutoSaveStatus("自动保存已开启", "saved");
    fillEditor(true);
    els.editorDrawer.classList.add("open");
    els.drawerBackdrop.classList.add("open");
    els.editorDrawer.setAttribute("aria-hidden", "false");
  }

  function closeEditor() {
    flushAutoSave();
    closeEditorCategoryMenu();
    closeTagSuggestions();
    state.editingId = null;
    els.editorDrawer.classList.remove("open");
    els.drawerBackdrop.classList.remove("open");
    els.editorDrawer.setAttribute("aria-hidden", "true");
  }

  function currentPrompt() {
    return state.prompts.find(p => p.id === state.editingId) || null;
  }

  function fillEditor(force) {
    const p = currentPrompt();
    if (!p) { closeEditor(); return; }
    if (force || document.activeElement !== els.editTitle) els.editTitle.value = p.title || "";
    if (force || !els.editCategorySelect.contains(document.activeElement)) renderEditorCategoryOptions(p.category || COMMON_CATEGORY);
    if (force || document.activeElement !== els.editTagInput) renderEditorTags(p.tags || "");
    if (force || document.activeElement !== els.editContent) els.editContent.value = p.content || "";
    if (force || document.activeElement !== els.editNotes) els.editNotes.value = p.notes || "";
    els.charCount.textContent = `${els.editContent.value.length} 字符`;
    els.createdAt.textContent = formatDate(p.createdAt);
    els.updatedAt.textContent = formatDate(p.updatedAt);
    els.useCount.textContent = `${p.useCount || 0} 次`;
    els.editFavoriteBtn.classList.toggle("active", !!p.isFavorite);
    els.drawerState.textContent = p.category ? `${p.category} · 编辑提示词` : "编辑提示词";
  }

  function setAutoSaveStatus(text, mode = "") {
    if (!els.autoSaveStatus) return;
    els.autoSaveStatus.textContent = text;
    els.autoSaveStatus.dataset.mode = mode;
  }

  function collectEditorPrompt() {
    const p = currentPrompt();
    if (!p) return null;
    return {
      ...p,
      title: els.editTitle.value.trim() || "未命名提示词",
      category: els.editCategory.value.trim() || COMMON_CATEGORY,
      tags: els.editTags.value.trim(),
      content: els.editContent.value,
      notes: els.editNotes.value
    };
  }

  function applyEditorPromptLocally(prompt) {
    const p = currentPrompt();
    if (!p || !prompt) return;
    Object.assign(p, prompt);
    p.updatedAt = new Date().toISOString();
    els.updatedAt.textContent = formatDate(p.updatedAt);
    els.drawerState.textContent = `${p.category || COMMON_CATEGORY} · 编辑提示词`;
  }

  function scheduleAutoSave() {
    if (!currentPrompt()) return;
    autoSaveDirty = true;
    clearTimeout(autoSaveTimer);
    setAutoSaveStatus("有未保存修改", "dirty");
    autoSaveTimer = setTimeout(flushAutoSave, AUTO_SAVE_DELAY);
  }

  function flushAutoSave() {
    clearTimeout(autoSaveTimer);
    autoSaveTimer = null;
    if (!autoSaveDirty) return;
    const prompt = collectEditorPrompt();
    if (!prompt) return;
    applyEditorPromptLocally(prompt);
    autoSaveDirty = false;
    setAutoSaveStatus("正在自动保存…", "saving");
    post("autoSavePrompt", { prompt });
    renderGrid();
  }

  function saveEditor(closeAfter = false) {
    clearTimeout(autoSaveTimer);
    autoSaveTimer = null;
    const prompt = collectEditorPrompt();
    if (!prompt) return;
    applyEditorPromptLocally(prompt);
    autoSaveDirty = false;
    setAutoSaveStatus("正在保存…", "saving");
    post("savePrompt", { prompt });
    if (closeAfter) closeEditor();
    else toast("修改已保存", "success");
  }

  function openTaxonomy() {
    state.newCategoryColor = nextUnusedCategoryColor();
    renderTaxonomy();
    els.taxonomyModal.classList.add("open");
    els.taxonomyModal.setAttribute("aria-hidden", "false");
    setTimeout(() => els.newCategoryInput.focus(), 0);
  }

  function closeTaxonomy() {
    els.taxonomyModal.classList.remove("open");
    els.taxonomyModal.setAttribute("aria-hidden", "true");
  }

  function renderTaxonomy() {
    const categories = categoriesWithCounts();
    const tags = tagNames();
    els.categoryManageCount.textContent = categories.length;
    els.tagManageCount.textContent = tags.length;
    renderNewCategoryPalette();

    els.categoryManageList.innerHTML = categories.map(([name, count]) => `
      <div class="taxonomy-row category-taxonomy-row" draggable="true" data-value="${encodeURIComponent(name)}">
        <span class="taxonomy-grip" title="拖动排序">⠿</span>
        <button class="taxonomy-color" type="button" data-color-category="${encodeURIComponent(name)}" style="--category-color:${categoryColor(name)}" title="设置分类颜色" aria-label="设置 ${safe(name)} 的颜色"></button>
        <span class="taxonomy-name">${safe(name)}</span>
        <span class="taxonomy-usage">${count} 条</span>
        <button class="taxonomy-rename" type="button" data-rename-category="${encodeURIComponent(name)}" title="重命名分类">${icons.edit}</button>
        <button class="taxonomy-delete" type="button" data-delete-category="${encodeURIComponent(name)}" title="删除分类">${icons.trash}</button>
      </div>`).join("");

    const tagCounts = new Map();
    state.prompts.forEach(p => splitTags(p.tags).forEach(t => tagCounts.set(t.toLocaleLowerCase(), (tagCounts.get(t.toLocaleLowerCase()) || 0) + 1)));
    els.tagManageList.innerHTML = tags.map(name => `
      <div class="taxonomy-row" draggable="true" data-value="${encodeURIComponent(name)}">
        <span class="taxonomy-grip" title="拖动排序">⠿</span>
        <span class="taxonomy-name">${safe(name)}</span>
        <span class="taxonomy-usage">${tagCounts.get(name.toLocaleLowerCase()) || 0} 条</span>
        <button class="taxonomy-rename" type="button" data-rename-tag="${encodeURIComponent(name)}" title="重命名标签">${icons.edit}</button>
        <button class="taxonomy-delete" type="button" data-delete-tag="${encodeURIComponent(name)}" title="删除标签">${icons.trash}</button>
      </div>`).join("");

    els.categoryManageList.querySelectorAll("[data-color-category]").forEach(btn => btn.addEventListener("click", e => {
      e.stopPropagation();
      openColorDialog(decodeURIComponent(btn.dataset.colorCategory || ""));
    }));

    els.categoryManageList.querySelectorAll("[data-rename-category]").forEach(btn => btn.addEventListener("click", async () => {
      const oldName = decodeURIComponent(btn.dataset.renameCategory || "");
      if (!oldName) return;
      const raw = await promptDialog({ title: "重命名分类", message: "输入新的分类名称。相关提示词会同步更新。", value: oldName, confirmText: "保存" });
      if (raw === null) return;
      const newName = raw.trim();
      if (!newName || newName === oldName) return;
      if (categoryNames().some(x => x.toLocaleLowerCase() === newName.toLocaleLowerCase() && x.toLocaleLowerCase() !== oldName.toLocaleLowerCase())) { toast("该分类已存在"); return; }
      if (state.category === oldName) state.category = newName;
      post("renameCategory", { oldName, newName });
    }));

    els.tagManageList.querySelectorAll("[data-rename-tag]").forEach(btn => btn.addEventListener("click", async () => {
      const oldName = decodeURIComponent(btn.dataset.renameTag || "");
      if (!oldName) return;
      const raw = await promptDialog({ title: "重命名标签", message: "输入新的标签名称。所有引用该标签的提示词会同步更新。", value: oldName, confirmText: "保存" });
      if (raw === null) return;
      const newName = raw.trim();
      if (!newName || newName === oldName) return;
      if (tagNames().some(x => x.toLocaleLowerCase() === newName.toLocaleLowerCase() && x.toLocaleLowerCase() !== oldName.toLocaleLowerCase())) { toast("该标签已存在"); return; }
      post("renameTag", { oldName, newName });
    }));

    els.categoryManageList.querySelectorAll("[data-delete-category]").forEach(btn => btn.addEventListener("click", async () => {
      const name = decodeURIComponent(btn.dataset.deleteCategory || "");
      const count = state.prompts.filter(p => p.category === name).length;
      const ok = await confirmDialog({
        title: "删除分类",
        message: count ? `确定删除“${name}”吗？\n其中 ${count} 条提示词会自动移动到其他可用分类。` : `确定删除“${name}”吗？\n该分类目前没有提示词。`,
        confirmText: "删除",
        kind: "danger"
      });
      if (ok) {
        if (state.category === name) state.category = "全部分类";
        post("deleteCategory", { name });
      }
    }));

    els.tagManageList.querySelectorAll("[data-delete-tag]").forEach(btn => btn.addEventListener("click", async () => {
      const name = decodeURIComponent(btn.dataset.deleteTag || "");
      const count = state.prompts.filter(p => splitTags(p.tags).some(t => t.toLocaleLowerCase() === name.toLocaleLowerCase())).length;
      const ok = await confirmDialog({ title: "删除标签", message: `确定删除“${name}”吗？\n它将从 ${count} 条提示词中同步移除。`, confirmText: "删除", kind: "danger" });
      if (ok) post("deleteTag", { name });
    }));

    wireTaxonomyDrag(els.categoryManageList, "reorderCategories");
    wireTaxonomyDrag(els.tagManageList, "reorderTags");
  }

  function wireTaxonomyDrag(container, messageType) {
    let dragged = null;
    container.querySelectorAll(".taxonomy-row").forEach(row => {
      row.addEventListener("dragstart", e => {
        if (row.draggable === false) { e.preventDefault(); return; }
        dragged = row;
        row.classList.add("dragging");
        e.dataTransfer.effectAllowed = "move";
        e.dataTransfer.setData("text/plain", row.dataset.value || "");
      });
      row.addEventListener("dragover", e => {
        if (!dragged || dragged === row) return;
        e.preventDefault();
        row.classList.add("drag-over");
      });
      row.addEventListener("dragleave", () => row.classList.remove("drag-over"));
      row.addEventListener("drop", e => {
        if (!dragged || dragged === row) return;
        e.preventDefault();
        row.classList.remove("drag-over");
        const rect = row.getBoundingClientRect();
        if (e.clientY > rect.top + rect.height / 2) row.after(dragged);
        else row.before(dragged);
        const items = [...container.querySelectorAll(".taxonomy-row")].map(x => decodeURIComponent(x.dataset.value || ""));
        if (messageType === "reorderCategories") state.categories = items;
        else state.tags = items;
        post(messageType, { items });
      });
      row.addEventListener("dragend", () => {
        row.classList.remove("dragging");
        container.querySelectorAll(".drag-over").forEach(x => x.classList.remove("drag-over"));
        dragged = null;
      });
    });
  }

  function addCategory() {
    const name = els.newCategoryInput.value.trim();
    if (!name) return;
    if (categoryNames().some(x => x.toLocaleLowerCase() === name.toLocaleLowerCase())) { toast("该分类已存在"); return; }
    els.newCategoryInput.value = "";
    post("addCategory", { name, color: state.newCategoryColor });
  }

  function addTag() {
    const name = els.newTagInput.value.trim();
    if (!name) return;
    if (tagNames().some(x => x.toLocaleLowerCase() === name.toLocaleLowerCase())) { toast("该标签已存在"); return; }
    els.newTagInput.value = "";
    post("addTag", { name });
  }

  function renderDataSettings(forceInput = false) {
    if (forceInput || document.activeElement !== els.dataDirectoryInput) {
      els.dataDirectoryInput.value = state.dataDirectory || "";
    }
    els.dataDirectoryMode.textContent = state.customDataDirectory ? "自定义目录" : "默认目录";
    els.dataDirectoryMode.classList.toggle("custom", !!state.customDataDirectory);
    els.dataFilePath.textContent = state.dataPath || "—";
    els.dataFilePath.title = state.dataPath || "";
    els.resetDataPathBtn.disabled = !state.customDataDirectory;
  }

  function openDataSettings() {
    renderDataSettings(true);
    els.dataModal.classList.add("open");
    els.dataModal.setAttribute("aria-hidden", "false");
    setTimeout(() => els.dataDirectoryInput.focus(), 80);
  }

  function closeDataSettings() {
    els.dataModal.classList.remove("open");
    els.dataModal.setAttribute("aria-hidden", "true");
  }

  function applyDataDirectory() {
    const path = els.dataDirectoryInput.value.trim();
    if (!path) { toast("请输入数据目录路径", "error"); return; }
    post("setDataDirectory", { path });
  }

  let dialogResolver = null;

  function closeAppDialog(result) {
    if (!els.appDialog.classList.contains("open")) return;
    els.appDialog.classList.remove("open");
    els.appDialog.setAttribute("aria-hidden", "true");
    const resolve = dialogResolver;
    dialogResolver = null;
    setTimeout(() => {
      els.appDialogInput.classList.remove("visible");
      els.appDialogInput.value = "";
    }, 160);
    if (resolve) resolve(result);
  }

  function openAppDialog({ title, message, confirmText = "确定", cancelText = "取消", kind = "default", input = false, value = "", placeholder = "" }) {
    if (dialogResolver) closeAppDialog(false);
    els.appDialogPanel.dataset.kind = kind;
    els.appDialogIcon.textContent = kind === "danger" ? "!" : kind === "warning" ? "!" : input ? "✎" : "?";
    els.appDialogTitle.textContent = title || "确认操作";
    els.appDialogMessage.textContent = message || "请确认是否继续。";
    els.appDialogConfirm.textContent = confirmText;
    els.appDialogCancel.textContent = cancelText;
    els.appDialogInput.classList.toggle("visible", !!input);
    els.appDialogInput.value = input ? String(value ?? "") : "";
    els.appDialogInput.placeholder = placeholder || "";
    els.appDialog.classList.add("open");
    els.appDialog.setAttribute("aria-hidden", "false");
    setTimeout(() => {
      if (input) { els.appDialogInput.focus(); els.appDialogInput.select(); }
      else els.appDialogConfirm.focus();
    }, 30);
    return new Promise(resolve => { dialogResolver = resolve; });
  }

  async function confirmDialog(options) {
    return !!(await openAppDialog({ ...options, input: false }));
  }

  async function promptDialog(options) {
    const result = await openAppDialog({ ...options, input: true });
    return result === false ? null : String(result ?? "");
  }

  els.appDialogConfirm.addEventListener("click", () => {
    closeAppDialog(els.appDialogInput.classList.contains("visible") ? els.appDialogInput.value : true);
  });
  els.appDialogCancel.addEventListener("click", () => closeAppDialog(false));
  els.appDialog.addEventListener("click", e => { if (e.target === els.appDialog) closeAppDialog(false); });
  els.appDialogInput.addEventListener("keydown", e => {
    if (e.key === "Enter") { e.preventDefault(); closeAppDialog(els.appDialogInput.value); }
  });

  function toast(text, kind = "") {
    const div = document.createElement("div");
    div.className = `toast ${kind}`;
    div.textContent = text;
    els.toastStack.appendChild(div);
    setTimeout(() => div.remove(), 2400);
  }

  document.querySelectorAll(".nav-item[data-scope]").forEach(btn => btn.addEventListener("click", () => {
    state.scope = btn.dataset.scope;
    state.category = "全部分类";
    render();
  }));

  els.searchInput.addEventListener("input", () => { state.search = els.searchInput.value; renderGrid(); });
  els.newPromptBtn.addEventListener("click", () => post("createPrompt"));
  els.emptyCreateBtn.addEventListener("click", () => post("createPrompt"));
  els.themeBtn.addEventListener("click", () => {
    state.isDarkTheme = !state.isDarkTheme;
    applyTheme();
    post("setTheme", { isDark: state.isDarkTheme });
  });
  els.trayBtn.addEventListener("click", () => post("hideToTray"));
  els.importBtn.addEventListener("click", () => post("import"));
  els.exportBtn.addEventListener("click", () => post("export"));
  els.folderBtn.addEventListener("click", openDataSettings);
  els.githubBtn?.addEventListener("click", () => post("openGitHub"));
  els.manageTaxonomyBtn.addEventListener("click", openTaxonomy);
  els.closeTaxonomyBtn.addEventListener("click", closeTaxonomy);
  els.taxonomyModal.addEventListener("click", e => { if (e.target === els.taxonomyModal) closeTaxonomy(); });
  els.closeDataBtn.addEventListener("click", closeDataSettings);
  els.dataModal.addEventListener("click", e => { if (e.target === els.dataModal) closeDataSettings(); });
  els.applyDataPathBtn.addEventListener("click", applyDataDirectory);
  els.chooseDataPathBtn.addEventListener("click", () => post("chooseDataDirectory"));
  els.openDataPathBtn.addEventListener("click", () => post("openDataFolder"));
  els.resetDataPathBtn.addEventListener("click", () => post("resetDataDirectory"));
  els.dataDirectoryInput.addEventListener("keydown", e => { if (e.key === "Enter") applyDataDirectory(); });
  els.closeColorDialogBtn.addEventListener("click", closeColorDialog);
  els.colorDialog.addEventListener("click", e => { if (e.target === els.colorDialog) closeColorDialog(); });
  els.addCategoryBtn.addEventListener("click", addCategory);
  els.addTagBtn.addEventListener("click", addTag);
  els.newCategoryInput.addEventListener("keydown", e => { if (e.key === "Enter") addCategory(); });
  els.newTagInput.addEventListener("keydown", e => { if (e.key === "Enter") addTag(); });
  els.closeDrawerBtn.addEventListener("click", closeEditor);
  els.drawerBackdrop.addEventListener("click", closeEditor);
  document.addEventListener("pointerdown", e => { if (els.editCategorySelect && !els.editCategorySelect.contains(e.target)) closeEditorCategoryMenu(); });
  els.editTitle.addEventListener("input", scheduleAutoSave);
  els.editCategoryTrigger.addEventListener("click", () => {
    const open = els.editCategorySelect.classList.toggle("open");
    els.editCategoryTrigger.setAttribute("aria-expanded", open ? "true" : "false");
  });
  els.editCategoryTrigger.addEventListener("keydown", e => {
    if (e.key === "Enter" || e.key === " " || e.key === "ArrowDown") { e.preventDefault(); els.editCategorySelect.classList.add("open"); els.editCategoryTrigger.setAttribute("aria-expanded", "true"); els.editCategoryMenu.querySelector(".selected")?.focus(); }
    if (e.key === "Escape") closeEditorCategoryMenu();
  });
  els.editTagInput.addEventListener("keydown", e => {
    const suggestions = getTagSuggestions();
    if (e.key === "ArrowDown" && suggestions.length) {
      e.preventDefault();
      tagSuggestionIndex = Math.min(tagSuggestionIndex + 1, suggestions.length - 1);
      renderTagSuggestions();
    } else if (e.key === "ArrowUp" && suggestions.length) {
      e.preventDefault();
      tagSuggestionIndex = tagSuggestionIndex <= 0 ? suggestions.length - 1 : tagSuggestionIndex - 1;
      renderTagSuggestions();
    } else if (e.key === "Enter") {
      e.preventDefault();
      if (tagSuggestionIndex >= 0 && suggestions[tagSuggestionIndex]) selectTagSuggestion(suggestions[tagSuggestionIndex]);
      else commitEditorTag();
    } else if (e.key === "Escape" && els.editTagSuggestions.classList.contains("open")) {
      e.preventDefault();
      e.stopPropagation();
      closeTagSuggestions();
    } else if (e.key === "Backspace" && !els.editTagInput.value) {
      const tags = editorTagValues();
      if (tags.length) { tags.pop(); renderEditorTags(tags.join(", ")); scheduleAutoSave(); renderTagSuggestions(); }
    }
  });
  els.editTagInput.addEventListener("input", () => { tagSuggestionIndex = -1; renderTagSuggestions(); });
  els.editTagInput.addEventListener("focus", () => { els.tagInputShell.classList.add("focused"); renderTagSuggestions(); });
  els.editTagInput.addEventListener("blur", () => { els.tagInputShell.classList.remove("focused"); setTimeout(closeTagSuggestions, 120); });
  els.tagInputShell.addEventListener("click", e => { if (!e.target.closest(".tag-token button") && !e.target.closest(".tag-suggestion-item")) els.editTagInput.focus(); });
  els.editContent.addEventListener("input", () => { els.charCount.textContent = `${els.editContent.value.length} 字符`; scheduleAutoSave(); });
  els.editNotes.addEventListener("input", scheduleAutoSave);
  els.editFavoriteBtn.addEventListener("click", () => {
    const p = currentPrompt();
    if (p) post("toggleFavorite", { id: p.id });
  });
  els.copyBtn.addEventListener("click", () => {
    const p = currentPrompt();
    if (!p) return;
    flushAutoSave();
    post("copyPrompt", { id: p.id });
    toast("标题和正文已复制到剪贴板", "success");
  });
  els.duplicateBtn.addEventListener("click", () => {
    const p = currentPrompt();
    if (p) { flushAutoSave(); post("duplicatePrompt", { id: p.id }); toast("已创建副本", "success"); }
  });
  els.deleteBtn.addEventListener("click", async () => {
    const p = currentPrompt();
    if (!p) return;
    const ok = await confirmDialog({ title: "删除提示词", message: `确定删除“${p.title}”吗？\n删除后无法撤销。`, confirmText: "删除", kind: "danger" });
    if (ok) {
      post("deletePrompt", { id: p.id });
      closeEditor();
    }
  });

  document.addEventListener("keydown", e => {
    if (e.ctrlKey && !e.shiftKey && e.key.toLowerCase() === "n") { e.preventDefault(); post("createPrompt"); }
    if (e.ctrlKey && !e.shiftKey && e.key.toLowerCase() === "f") { e.preventDefault(); els.searchInput.focus(); els.searchInput.select(); }
    if (e.ctrlKey && !e.shiftKey && e.key.toLowerCase() === "s") { e.preventDefault(); if (autoSaveDirty) flushAutoSave(); else toast("当前内容已自动保存"); }
    if (e.key === "Escape") {
      if (els.appDialog.classList.contains("open")) closeAppDialog(false);
      else if (els.colorDialog.classList.contains("open")) closeColorDialog();
      else if (els.taxonomyModal.classList.contains("open")) closeTaxonomy();
      else if (els.dataModal.classList.contains("open")) closeDataSettings();
      else if (state.editingId) closeEditor();
    }
  });

  function handleHostMessage(msg) {
    msg = msg || {};
    if (msg.type === "state") {
      const payload = msg.payload || {};
      const incoming = payload.state || {};
      state.prompts = (incoming.prompts || incoming.Prompts || []).map(normalizePrompt);
      state.categories = categoryOrder(incoming.categories || incoming.Categories || []);
      state.categoryColors = { ...(incoming.categoryColors || incoming.CategoryColors || {}) };
      state.categoryPalette = Array.isArray(payload.categoryPalette) && payload.categoryPalette.length ? payload.categoryPalette.slice(0, 20) : [...DEFAULT_CATEGORY_PALETTE];
      state.tags = distinct(incoming.tags || incoming.Tags || []);
      state.isDarkTheme = incoming.isDarkTheme ?? incoming.IsDarkTheme ?? false;
      state.dataPath = payload.dataPath || "";
      state.dataDirectory = payload.dataDirectory || "";
      state.defaultDataDirectory = payload.defaultDataDirectory || "";
      state.customDataDirectory = !!payload.customDataDirectory;
      state.prompts.sort((a, b) => a.sortOrder - b.sortOrder);

      if (state.category !== "全部分类" && !categoryNames().includes(state.category)) state.category = "全部分类";

      if (payload.focusId) {
        state.editingId = String(payload.focusId);
        render();
        openEditor(state.editingId);
      } else {
        render();
      }

      if (payload.reason === "deleted") toast("提示词已删除");
      if (payload.reason === "imported") toast(`已导入 ${state.prompts.length} 条提示词`, "success");
      if (payload.reason === "saved") setAutoSaveStatus("已保存", "saved");
      if (payload.reason === "categoryAdded") { state.newCategoryColor = nextUnusedCategoryColor(); renderNewCategoryPalette(); toast("分类已添加", "success"); }
      if (payload.reason === "categoryDeleted") toast("分类已删除，相关提示词已移动到其他可用分类", "success");
      if (payload.reason === "categoryRenamed") toast("分类已重命名", "success");
      if (payload.reason === "categoryColorChanged") toast("分类颜色已更新", "success");
      if (payload.reason === "tagDeleted") { if (state.editingId) { fillEditor(true); renderTagSuggestions(); } toast("标签已删除并从所有提示词中彻底移除", "success"); }
      if (payload.reason === "tagRenamed") { if (state.editingId) { fillEditor(true); renderTagSuggestions(); } toast("标签已重命名并同步到提示词", "success"); }
    }
    if (msg.type === "confirmImport") {
      const count = Number(msg.payload?.count || 0);
      confirmDialog({
        title: "导入提示词库",
        message: `将导入 ${count} 条提示词，并替换当前提示词库。\n建议先导出备份后再继续。`,
        confirmText: "导入并替换",
        kind: "warning"
      }).then(ok => post(ok ? "confirmImport" : "cancelImport"));
    }
    if (msg.type === "confirmDataDirectory") {
      const path = String(msg.payload?.path || "");
      confirmDialog({
        title: "切换数据目录",
        message: `目标目录已经存在 prompts.json。\n确认后将切换并加载该目录中的提示词库：\n${path}`,
        confirmText: "切换并加载",
        kind: "warning"
      }).then(ok => post(ok ? "confirmDataDirectory" : "cancelDataDirectory"));
    }
    if (msg.type === "autoSaveAck") {
      const id = String(msg.payload?.id || "");
      const p = state.prompts.find(x => x.id === id);
      if (p && msg.payload?.updatedAt) p.updatedAt = msg.payload.updatedAt;
      if (id && state.editingId === id) {
        if (msg.payload?.updatedAt) els.updatedAt.textContent = formatDate(msg.payload.updatedAt);
        setAutoSaveStatus("已自动保存", "saved");
      }
    }
    if (msg.type === "toast") toast(msg.payload?.text || "", msg.payload?.kind || "");
  }

  async function waitForBridge() {
    for (let i = 0; i < 100; i++) {
      if (window.go?.main?.App?.HandleMessage) {
        await post("ready");
        return;
      }
      await new Promise(resolve => setTimeout(resolve, 50));
    }
    toast("Wails 后端未就绪，请重新启动 PromptNest。", "error");
  }

  waitForBridge();
}
