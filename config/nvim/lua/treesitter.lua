-- Treesitter setup
require("nvim-treesitter").setup()

-- Install parsers if missing
local wanted = {
  "bash", "c", "css", "diff", "go", "gomod", "gosum", "html",
  "javascript", "json", "lua", "luadoc", "markdown", "markdown_inline",
  "python", "query", "regex", "toml", "tsx", "typescript", "vim",
  "vimdoc", "yaml",
}

local installed = {}
for _, lang in ipairs(require("nvim-treesitter").get_installed()) do
  installed[lang] = true
end

local missing = {}
for _, lang in ipairs(wanted) do
  if not installed[lang] then
    table.insert(missing, lang)
  end
end

if #missing > 0 then
  require("nvim-treesitter").install(missing)
end

-- Textobjects
require("nvim-treesitter-textobjects").setup({
  select = {
    enable = true,
    lookahead = true,
    keymaps = {
      ["af"] = "@function.outer",
      ["if"] = "@function.inner",
      ["ac"] = "@class.outer",
      ["ic"] = "@class.inner",
      ["aa"] = "@parameter.outer",
      ["ia"] = "@parameter.inner",
    },
  },
  move = {
    enable = true,
    goto_next_start = {
      ["]f"] = "@function.outer",
      ["]c"] = "@class.outer",
      ["]a"] = "@parameter.inner",
    },
    goto_next_end = {
      ["]F"] = "@function.outer",
      ["]C"] = "@class.outer",
    },
    goto_previous_start = {
      ["[f"] = "@function.outer",
      ["[c"] = "@class.outer",
      ["[a"] = "@parameter.inner",
    },
    goto_previous_end = {
      ["[F"] = "@function.outer",
      ["[C"] = "@class.outer",
    },
  },
  swap = {
    enable = true,
    swap_next = { ["<leader>a"] = "@parameter.inner" },
    swap_previous = { ["<leader>A"] = "@parameter.inner" },
  },
})

-- Auto-close HTML/JSX tags
require("nvim-ts-autotag").setup()

-- Treesitter-aware commenting
require("ts-comments").setup()
