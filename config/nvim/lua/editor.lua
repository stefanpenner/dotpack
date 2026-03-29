-- Git signs
require("gitsigns").setup({
  signs = {
    add = { text = "▎" },
    change = { text = "▎" },
    delete = { text = "" },
    topdelete = { text = "" },
    changedelete = { text = "▎" },
    untracked = { text = "▎" },
  },
  on_attach = function(buffer)
    local gs = require("gitsigns")
    local map = function(mode, l, r, desc)
      vim.keymap.set(mode, l, r, { buffer = buffer, desc = desc })
    end
    map("n", "]h", function() gs.nav_hunk("next") end, "Next hunk")
    map("n", "[h", function() gs.nav_hunk("prev") end, "Prev hunk")
    map("n", "]H", function() gs.nav_hunk("last") end, "Last hunk")
    map("n", "[H", function() gs.nav_hunk("first") end, "First hunk")
    map({ "n", "v" }, "<leader>ghs", ":Gitsigns stage_hunk<CR>", "Stage hunk")
    map({ "n", "v" }, "<leader>ghr", ":Gitsigns reset_hunk<CR>", "Reset hunk")
    map("n", "<leader>ghS", gs.stage_buffer, "Stage buffer")
    map("n", "<leader>ghu", gs.undo_stage_hunk, "Undo stage hunk")
    map("n", "<leader>ghR", gs.reset_buffer, "Reset buffer")
    map("n", "<leader>ghp", gs.preview_hunk_inline, "Preview hunk inline")
    map("n", "<leader>ghb", function() gs.blame_line({ full = true }) end, "Blame line")
    map("n", "<leader>ghd", gs.diffthis, "Diff this")
  end,
})

-- Flash (enhanced search/jump)
require("flash").setup()

-- Grug-far (multi-file search & replace)
require("grug-far").setup()

-- Trouble (diagnostics list)
require("trouble").setup()

-- Todo comments
require("todo-comments").setup()

-- Which-key (keybinding discovery)
require("which-key").setup({
  preset = "modern",
  spec = {
    { "<leader>b", group = "buffer" },
    { "<leader>c", group = "code" },
    { "<leader>f", group = "find" },
    { "<leader>g", group = "git" },
    { "<leader>gh", group = "hunks" },
    { "<leader>q", group = "quit/session" },
    { "<leader>s", group = "search" },
    { "<leader>sn", group = "noice" },
    { "<leader>t", group = "test" },
    { "<leader>u", group = "ui" },
    { "<leader>w", group = "windows" },
    { "<leader>x", group = "diagnostics" },
    { "<leader><tab>", group = "tabs" },
    { "[", group = "prev" },
    { "]", group = "next" },
    { "g", group = "goto" },
  },
})

-- Mini.ai (smart text objects)
require("mini.ai").setup()

-- Mini.pairs (auto-close brackets)
require("mini.pairs").setup()

-- Persistence (session management)
require("persistence").setup()

-- Yanky (improved yank/paste)
require("yanky").setup({
  highlight = { timer = 200 },
})

-- Render markdown
require("render-markdown").setup()

-- Testing
require("neotest").setup({
  adapters = {
    require("neotest-golang"),
  },
})
