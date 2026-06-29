#!/usr/bin/env python3
"""Convert a pg_dump --schema-only file into a Camel `action: raw` YAML baseline.

Splits SQL statements with dollar-quote awareness ($$ / $_$ / $tag$) so plpgsql
function bodies survive intact, drops session-only noise (SET, set_config) and
comments, and emits each top-level statement as a YAML block scalar list item.
"""
import re
import sys

src, dst = sys.argv[1], sys.argv[2]
sql = open(src).read()

# --- dollar-quote-aware statement splitter -------------------------------
stmts = []
buf = []
i = 0
n = len(sql)
dollar = None  # active dollar tag, e.g. "$$" or "$_$"
tag_re = re.compile(r"\$[A-Za-z_]*\$")
at_line_start = True
instr = False  # inside a single-quoted string literal
while i < n:
    if dollar is None and not instr:
        # Skip whole-line "--" comments (they carry semicolons in pg_dump
        # headers). Only outside dollar quotes / strings, so plpgsql bodies
        # and string literals are untouched.
        if at_line_start and sql.startswith("--", i):
            j = sql.find("\n", i)
            i = n if j == -1 else j + 1
            continue
        m = tag_re.match(sql, i)
        if m:
            dollar = m.group(0)
            buf.append(dollar)
            i = m.end()
            at_line_start = False
            continue
        c = sql[i]
        at_line_start = c == "\n"
        if c == "'":
            instr = True
            buf.append(c)
            i += 1
            continue
        if c == ";":
            buf.append(";")
            stmts.append("".join(buf))
            buf = []
            i += 1
            continue
        buf.append(c)
        i += 1
    elif instr:
        at_line_start = False
        c = sql[i]
        if c == "'":
            if i + 1 < n and sql[i + 1] == "'":  # doubled '' escape
                buf.append("''")
                i += 2
                continue
            instr = False
        buf.append(c)
        i += 1
    else:
        at_line_start = False
        if sql.startswith(dollar, i):
            buf.append(dollar)
            i += len(dollar)
            dollar = None
            continue
        buf.append(sql[i])
        i += 1
if "".join(buf).strip():
    stmts.append("".join(buf))


def clean(stmt):
    # strip leading comment-only / blank lines, keep inner structure
    lines = stmt.split("\n")
    kept = []
    for ln in lines:
        s = ln.strip()
        if not kept and (s == "" or s.startswith("--")):
            continue
        kept.append(ln.rstrip())
    while kept and kept[-1].strip() == "":
        kept.pop()
    return "\n".join(kept).strip()


SKIP = re.compile(r"^\s*(SET\s|SELECT\s+pg_catalog\.set_config)", re.IGNORECASE)

cleaned = []
for s in stmts:
    c = clean(s)
    if not c or SKIP.match(c):
        continue
    cleaned.append(c)


def emit_block(out, statements, indent="      "):
    for st in statements:
        out.append(f"{indent}- |")
        for line in st.split("\n"):
            out.append(f"{indent}  {line}")


out = []
out.append("# Camel schema baseline for acme — generated from pg_dump --schema-only.")
out.append("# action: raw so native enum types, functions and triggers survive verbatim")
out.append("# (Camel's DSL cannot express PG enum types / plpgsql). New schema changes")
out.append("# going forward should be authored as readable Camel YAML migrations.")
out.append("up:")
out.append("  schema:")
out.append("    action: raw")
out.append("    statements:")
emit_block(out, cleaned)
out.append("down:")
out.append("  schema:")
out.append("    action: raw")
out.append("    statements:")
emit_block(out, ["DROP SCHEMA public CASCADE;", "CREATE SCHEMA public;"])
out.append("")

open(dst, "w").write("\n".join(out))
print(f"statements: {len(cleaned)}")
print(f"wrote: {dst}")
