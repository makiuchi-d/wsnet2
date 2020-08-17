using System;
using System.Text;
using System.Security.Cryptography;
using WSNet2.Core;

public static class WSNet2Helper
{
    public static byte[] Serialize<T>(T v) where T : class, IWSNetSerializable
    {
        // FIXME: もう少し便利な方法が提供してほしい
        var w = WSNet2.Core.Serialization.GetWriter();
        lock (w)
        {
            w.Write<T>(v);
            var seg = w.ArraySegment();
            var ret = new byte[seg.Count];
            System.Array.Copy(seg.Array, seg.Offset, ret, 0, seg.Count);
            return ret;
        }
    }

    public static byte[] Serialize(int v)
    {
        // FIXME: もう少し便利な方法が提供してほしい
        var w = WSNet2.Core.Serialization.GetWriter();
        lock (w)
        {
            w.Write(v);
            var seg = w.ArraySegment();
            var ret = new byte[seg.Count];
            System.Array.Copy(seg.Array, seg.Offset, ret, 0, seg.Count);
            return ret;
        }
    }

    public static byte[] Serialize(string v)
    {
        // FIXME: もう少し便利な方法が提供してほしい
        var w = WSNet2.Core.Serialization.GetWriter();
        lock (w)
        {
            w.Write(v);
            var seg = w.ArraySegment();
            var ret = new byte[seg.Count];
            System.Array.Copy(seg.Array, seg.Offset, ret, 0, seg.Count);
            return ret;
        }
    }

    public static AuthData GenAuthData(string key, string userid)
    {
        var auth = new AuthData();

        auth.Timestamp = ((DateTimeOffset)DateTime.UtcNow).ToUnixTimeSeconds().ToString();

        var rng = new RNGCryptoServiceProvider();
        var nbuf = new byte[8];
        rng.GetBytes(nbuf);
        auth.Nonce = BitConverter.ToString(nbuf).Replace("-", "").ToLower();

        var str = userid + auth.Timestamp + auth.Nonce;
        var hmac = new HMACSHA256(Encoding.ASCII.GetBytes(key));
        var hash = hmac.ComputeHash(Encoding.ASCII.GetBytes(str));
        auth.Hash = BitConverter.ToString(hash).Replace("-", "").ToLower();

        return auth;
    }

    public static void RegisterTypes()
    {
        Serialization.Register<SampleClient.StrMessage>(1);
        Serialization.Register<GameScript.SyncPositionMessage>(2);
        Serialization.Register<GameScript.EmptyMessage>(3);
        Serialization.Register<GameScript.PlayerMoveMessage>(4);
    }
}