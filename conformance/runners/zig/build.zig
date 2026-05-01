const std = @import("std");

pub fn build(b: *std.Build) void {
    const target = b.standardTargetOptions(.{});
    const optimize = b.standardOptimizeOption(.{});

    const asun_mod = b.createModule(.{
        .root_source_file = b.path("../../../asun-zig/src/asun.zig"),
        .target = target,
        .optimize = optimize,
    });

    const runner_mod = b.createModule(.{
        .root_source_file = b.path("src/main.zig"),
        .target = target,
        .optimize = optimize,
    });
    runner_mod.addImport("asun", asun_mod);

    const exe = b.addExecutable(.{
        .name = "asun-conformance-zig",
        .root_module = runner_mod,
    });

    const run_cmd = b.addRunArtifact(exe);
    if (b.args) |args| run_cmd.addArgs(args);

    const run_step = b.step("run", "Run ASUN Zig encode conformance");
    run_step.dependOn(&run_cmd.step);

    b.installArtifact(exe);
}
