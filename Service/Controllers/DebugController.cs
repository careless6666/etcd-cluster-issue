using System;
using System.Net.Http;
using System.Net.Security;
using System.Security.Cryptography.X509Certificates;
using System.Text;
using System.Threading.Tasks;
using Etcdserverpb;
using Google.Protobuf;
using Grpc.Core;
using Grpc.Net.Client;
using Grpc.Net.Client.Configuration;
using Microsoft.AspNetCore.Mvc;
using Microsoft.Extensions.DependencyInjection;
using Microsoft.Extensions.Logging;
using V3Lockpb;

namespace Service.Controllers;

[Microsoft.AspNetCore.Components.Route("api/v1/debug")]
public class DebugController : Controller
{
    private readonly IServiceProvider _serviceProvider;

    public DebugController(IServiceProvider serviceProvider)
    {
        _serviceProvider = serviceProvider;
    }

    [HttpGet("test")]
    public async Task Test()
    {
        
        var connectionString = "https://etcd-stg-3.companyName:2379";

        MethodConfig _defaultGrpcMethodConfig = new()
        {
            Names = { MethodName.Default },
            RetryPolicy = new RetryPolicy
            {
                MaxAttempts = 5,
                InitialBackoff = TimeSpan.FromSeconds(1),
                MaxBackoff = TimeSpan.FromSeconds(5),
                BackoffMultiplier = 1.5,
                RetryableStatusCodes = { Grpc.Core.StatusCode.Unavailable }
            }
        };

        RetryThrottlingPolicy _defaultRetryThrottlingPolicy = new()
        {
            MaxTokens = 10,
            TokenRatio = 0.1
        };
        var clientCert =
            "-----BEGIN CERTIFICATE-----\nMIIEFq\n-----END CERTIFICATE-----";
        var clientKey =
            "-----BEGIN RSA PRIVATE KEY-----\nMIIu22wlZ\n-----END RSA PRIVATE KEY-----";
        var clientCertificate = X509Certificate2.CreateFromPem(clientCert, clientKey);
        var caStr =
            "-----BEGIN CERTIFICATE-----\nMIIDzjCCAUNRCem\n-----END CERTIFICATE-----";
        var caCertificate = new X509Certificate(Encoding.UTF8.GetBytes(caStr));

        X509CertificateCollection collection = new X509Certificate2Collection();
        collection.Add(clientCertificate);
        collection.Add(caCertificate);

        var sslOptions = new SslClientAuthenticationOptions
        {
            // Leave certs unvalidated for debugging
            RemoteCertificateValidationCallback = delegate { return true; },
            ClientCertificates = collection,
        };

        var socketHandler = new SocketsHttpHandler
        {
            SslOptions = sslOptions,
        };

        var options = new GrpcChannelOptions
        {
            ServiceConfig = new ServiceConfig
            {
                MethodConfigs = { _defaultGrpcMethodConfig },
                RetryThrottling = _defaultRetryThrottlingPolicy // { new RoundRobinConfig() }
            },
            HttpHandler = socketHandler,
            Credentials = ChannelCredentials.SecureSsl,
            LoggerFactory = _serviceProvider.GetRequiredService<ILoggerFactory>(),
        };

        var channel = GrpcChannel.ForAddress(connectionString, options);
        var lockClient = new Lock.LockClient(channel);
        var leaseClient = new Lease.LeaseClient(channel);
        var lease = await leaseClient.LeaseGrantAsync(new LeaseGrantRequest { ID = new Random().NextInt64(), TTL = 15 });
        var lockRes = await lockClient.LockAsync(new LockRequest
            {
                Lease = lease.ID,
                Name = ByteString.CopyFromUtf8("/ic-me-daemon-global-sync/election")
            }
            , deadline: new DateTime(DateTime.UtcNow.Ticks, DateTimeKind.Utc).AddSeconds(10)
        );
        Console.WriteLine(lockRes);
    }
    
}