using Chat.Client.Core;
using Chat.Client.Core.Abstractions;
using CommunityToolkit.Mvvm.ComponentModel;
using CommunityToolkit.Mvvm.Input;
using System.Collections.ObjectModel;

namespace Chat.Client.ViewModels;

public partial class RoomsViewModel : ObservableObject
{
    private readonly IChatApi _api;
    private readonly IAuthService _auth;

    public RoomsViewModel(IChatApi api, IAuthService auth)
    {
        _api = api;
        _auth = auth;
    }

    public ObservableCollection<Room> Rooms { get; } = new();

    [ObservableProperty] private Room? selectedRoom;
    [ObservableProperty] private string status = "";

    partial void OnSelectedRoomChanged(Room? value)
    {
        if (value is null) return;
        MainThread.BeginInvokeOnMainThread(async () =>
        {
            await Shell.Current.GoToAsync($"chat?roomId={Uri.EscapeDataString(value.Id)}&roomName={Uri.EscapeDataString(value.Name)}");
            SelectedRoom = null;
        });
    }

    public async Task EnsureLoadedAsync()
    {
        if (Rooms.Count == 0) await ReloadAsync();
    }

    [RelayCommand]
    public async Task ReloadAsync()
    {
        try
        {
            Status = "Loading rooms...";
            Rooms.Clear();
            var rooms = await _api.GetRoomsAsync(CancellationToken.None);
            foreach (var r in rooms) Rooms.Add(r);
            Status = $"Rooms: {Rooms.Count}";
        }
        catch (Exception ex)
        {
            Status = $"Error: {ex.Message}";
        }
    }

    [RelayCommand]
    public async Task CreateRoomAsync()
    {
        try
        {
            var name = await Shell.Current.DisplayPromptAsync("Create room", "Room name?");
            if (string.IsNullOrWhiteSpace(name)) return;

            var room = await _api.CreateRoomAsync(name.Trim(), CancellationToken.None);
            Rooms.Insert(0, room);
            Status = $"Created: {room.Name}";
        }
        catch (Exception ex)
        {
            Status = $"Error: {ex.Message}";
        }
    }

    [RelayCommand]
    public async Task SignInAsync()
    {
        if (!await _auth.IsEnabledAsync())
        {
            Status = "Auth.Enabled=false (appsettings.json)";
            return;
        }

        await _auth.SignInAsync(CancellationToken.None);
        Status = "Signed in.";
    }

    [RelayCommand]
    public async Task SignOutAsync()
    {
        await _auth.SignOutAsync(CancellationToken.None);
        Status = "Signed out.";
    }
}