fn main() -> Result<(), Box<dyn std::error::Error>> {
    let proto_files = [
        "../../../proto/scheduler.proto",
        "../../../proto/executor.proto",
        "../../../proto/worker.proto",
        "../../../proto/durable_engine.proto",
    ];

    tonic_build::configure()
        .build_server(false)
        .build_client(true)
        .out_dir("src/proto")
        .compile(&proto_files, &["../../../proto"])?;

    for proto_file in proto_files.iter() {
        println!("cargo:rerun-if-changed={}", proto_file);
    }

    Ok(())
}
